package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/vault/api"

	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
)

type vaultKey struct {
	KeyPath    string `json:"key_path"`
	KeyType    string `json:"key_type"`
	PrivateKey string `json:"private_key"`
}

const (
	permFile   os.FileMode = 0o644
	permFolder os.FileMode = 0o755
)

var (
	fVaultAddr  = flag.String("vault-addr", "http://localhost:8200", "vault address")
	fVaultToken = flag.String("vault-token", "", "vault token")
	fVaultPath  = flag.String("vault-path", "iden3", "vault path to KV secrets")
	fDID        = flag.String("did", "", "did to export, if this is empty, all dids will be exported")
	fOperation  = flag.String("operation", "export", "operation to perform: export or import")
	fOutPutFile = flag.String("output-file", "keys.json", "output file")
	fInPutFile  = flag.String("input-file", "keys.json", "input file")
)

func main() {
	ctx, cancel := context.WithCancel(log.NewContext(context.Background(), -4, 1, os.Stdout))
	defer cancel()
	flag.Parse()
	if fVaultAddr == nil || fVaultToken == nil {
		log.Error(ctx, "vault-addr and vault-token are required")
		return
	}

	if *fVaultToken == "" {
		log.Error(ctx, "vault-token is required")
		return
	}

	vaultCfg := providers.Config{
		UserPassAuthEnabled: false,
		Address:             *fVaultAddr,
		Token:               *fVaultToken,
	}

	vaultCli, vaultErr := providers.VaultClient(ctx, vaultCfg)
	if vaultErr != nil {
		log.Error(ctx, "cannot initialize vault client", "err", vaultErr)
		return
	}

	switch *fOperation {
	case "export":
		if err := export(ctx, vaultCli, *fVaultPath, *fOutPutFile, fDID); err != nil {
			log.Error(ctx, "cannot export keys", "err", err)
			return
		}
	case "import":
		if err := importFn(ctx, vaultCli, *fInPutFile); err != nil {
			log.Error(ctx, "cannot import keys", "err", err)
			return
		}
	default:
		log.Error(ctx, "unknown operation", "operation", *fOperation)
	}
}

func importFn(ctx context.Context, vaultCli *api.Client, inputFile string) error {
	file, err := os.ReadFile(inputFile)
	if err != nil {
		log.Error(ctx, "cannot read file", "err", err)
		return err
	}
	writeKeys(ctx, vaultCli, file)
	return nil
}

func export(ctx context.Context, vaultCli *api.Client, vaultPath string, outputFolder string, did *string) error {
	vaultKeys := make([]*vaultKey, 0)
	walkIden3(ctx, vaultCli, vaultPath, "", func(srcKeyPath string) {
		vaultKey := readKey(ctx, vaultCli, srcKeyPath, did)
		if vaultKey != nil {
			vaultKeys = append(vaultKeys, vaultKey)
		}
	})

	dirPath := "./keys"
	if err := os.MkdirAll(dirPath, permFolder); err != nil {
		log.Error(ctx, "cannot create keys dir", "err", err)
		return err
	}

	file, err := json.MarshalIndent(vaultKeys, "", " ")
	if err != nil {
		log.Error(ctx, "cannot marshal keys", "err", err)
		return err
	}

	if err := os.WriteFile(dirPath+"/"+outputFolder, file, permFile); err != nil {
		log.Error(ctx, "cannot writeKeys file", "err", err)
		return err
	}

	log.Info(ctx, "keys exported", "keys", len(vaultKeys))
	return nil
}

// walkIden3 - walk iden3 keys in vault
func walkIden3(ctx context.Context, vaultCli *api.Client, mountPath, keysPath string, f func(string)) {
	s, err := vaultCli.Logical().List(path.Join(mountPath, "keys", keysPath))
	if err != nil {
		panic(err)
	}
	if s == nil || s.Data == nil {
		log.Info(ctx, "no keys found")
		return
	}
	keys, ok := s.Data["keys"].([]interface{})
	if !ok {
		log.Error(ctx, "unable to get keys for path", "path", keysPath)
		return
	}
	for _, k := range keys {
		key, ok := k.(string)
		if !ok {
			log.Error(ctx, "unable to get key", "key", k)
			continue
		}
		keyPath := path.Join(keysPath, key)
		if strings.HasSuffix(key, "/") {
			walkIden3(ctx, vaultCli, mountPath, keyPath, f)
		} else {
			f(keyPath)
		}
	}
}

// readKey - read key from vault
func readKey(ctx context.Context, srcVaultCli *api.Client, keyPath string, did *string) *vaultKey {
	if did != nil && *did != "" {
		if !strings.HasPrefix(keyPath, *did) {
			return nil
		}
	}
	s, err := srcVaultCli.Logical().Read(path.Join(*fVaultPath, "private", keyPath))
	if err != nil {
		log.Error(ctx, "cannot read key", "err", err, "keyPath", keyPath)
		return nil
	}
	if s == nil {
		log.Error(ctx, "no data in src key"+keyPath)
		return nil
	}

	data := make(map[string]any)

	keyType, ok := s.Data["key_type"].(string)
	if !ok {
		log.Error(ctx, "cannot get key_type", "keyPath", keyPath)
		return nil
	}
	data["key_type"] = keyType

	privateKey, ok := s.Data["private_key"].(string)
	if !ok {
		log.Error(ctx, "cannot get private_key", "keyPath", keyPath)
		return nil
	}
	data["private_key"] = privateKey

	for k, v := range s.Data {
		if k == "key_type" || k == "private_key" || k == "public_key" {
			continue
		}
		data[k] = v
	}
	vaultKey := &vaultKey{
		KeyPath:    keyPath,
		KeyType:    keyType,
		PrivateKey: privateKey,
	}
	return vaultKey
}

// writeKeys - write keys to vault
func writeKeys(ctx context.Context, vaultCli *api.Client, file []byte) {
	var keys []vaultKey
	if err := json.Unmarshal(file, &keys); err != nil {
		log.Error(ctx, "cannot unmarshal file", "err", err)
		return
	}
	importedKeys := 0
	for _, key := range keys {
		keyPath := key.KeyPath
		data := make(map[string]any)
		data["key_type"] = key.KeyType
		data["private_key"] = key.PrivateKey
		_, err := vaultCli.Logical().Write(path.Join(*fVaultPath, "import", keyPath), data)
		if err != nil {
			log.Error(ctx, "cannot write key", "err", err, "keyPath", keyPath)
			continue
		}
		importedKeys++
	}

	log.Info(ctx, "keys imported", "keys", importedKeys)
}
