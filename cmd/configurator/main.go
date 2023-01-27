package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/log"
)

const tip = "tip"

func main() {
	logBuffer := bytes.Buffer{}
	ctx := log.NewContext(context.Background(), log.LevelInfo, log.OutputText, &logBuffer)
	defer func() {
		log.Debug(ctx, "configurator finished")
		os.Stderr.WriteString("\r\n")
		os.Stderr.Write(logBuffer.Bytes())
	}()

	log.Info(ctx, "Configurator started")
	template := flag.String("template", "config.toml.sample", "a string")
	output := flag.String("output", "config.toml.new", "a string")
	flag.Parse()
	if len(os.Args) != 3 {
		fmt.Println("Usage: configurator <template-file> <output-file>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := guardInputs(*template, *output); err != nil {
		fmt.Println(err.Error())
		log.Error(ctx, "preconditions", err)
		return
	}
	in, err := os.Open(*template)
	if err != nil {
		fmt.Printf("Cannot open templates file <%s>: %v", *template, err)
		log.Error(ctx, "cannot open template file", err)
		return
	}
	defer in.Close()

	out := bytes.Buffer{}
	if err := configurator(ctx, in, &out); err != nil {
		log.Error(ctx, "tip", err)
		return
	}
	fp, err := os.Create(*output)
	if err != nil {
		fmt.Println("Error creating file configuration.")
		log.Error(ctx, "cannot open output file", err)
		return
	}

	_, err = fp.Write(out.Bytes())
	defer fp.Close()
	if err != nil {
		fmt.Println("Error writing file configuration.")
		log.Error(ctx, "cannot write in output file", err)
	}
}

func guardInputs(template, output string) error {
	if _, err := os.Stat(template); err != nil {
		return fmt.Errorf("cannot read template config file <%s>: %w", template, err)
	}
	if _, err := os.Stat(output); err == nil {
		return fmt.Errorf("cannot overwrite existing file <%s>", output)
	}
	return nil
}

func configurator(ctx context.Context, tpl io.Reader, output io.Writer) error {
	conf, err := defaultConfiguration(tpl)
	if err != nil {
		return err
	}

	if err := askForConfiguration(ctx, conf); err != nil {
		return err
	}

	out, err := toml.Marshal(conf)
	if err != nil {
		return err
	}

	_, err = output.Write(out)
	return err
}

func askForConfiguration(ctx context.Context, defaults *config.Configuration) error {
	t := reflect.TypeOf(*defaults)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		description := field.Tag.Get(tip)
		defaul := reflect.ValueOf(defaults).Elem().Field(i)
		switch field.Type.String() {
		case "config.Database":
			if err := askForSectionConfiguration(&defaults.Database); err != nil {
				return err
			}
		case "config.KeyStore":
			if err := askForSectionConfiguration(&defaults.KeyStore); err != nil {
				return err
			}
		case "config.Log":
			if err := askForSectionConfiguration(&defaults.Log); err != nil {
				return err
			}
		case "config.Ethereum":
			if err := askForSectionConfiguration(&defaults.Ethereum); err != nil {
				return err
			}
		case "config.Circuit":
			if err := askForSectionConfiguration(&defaults.Circuit); err != nil {
				return err
			}
		case "config.ReverseHashService":
			if err := askForSectionConfiguration(&defaults.ReverseHashService); err != nil {
				return err
			}
		case "config.Prover":
			if err := askForSectionConfiguration(&defaults.Prover); err != nil {
				return err
			}
		default:
			if err := setBasicType(reflect.ValueOf(defaults).Elem().Field(i), description, fieldName, defaul, field.Type.String()); err != nil {
				log.Error(ctx, "Unknown section", err, "section", field.Type.String())
				return err
			}
		}
	}
	return nil
}

type section interface {
	config.Database | config.Circuit | config.Log | config.ReverseHashService | config.Ethereum | config.KeyStore | config.Prover
}

func askForSectionConfiguration[T section](defaults *T) error {
	t := reflect.TypeOf(*defaults)

	fmt.Println(sub(fmt.Sprintf("\n\n%s Configuration:", t.Name())))
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		description := field.Tag.Get(tip)
		defaul := reflect.ValueOf(defaults).Elem().Field(i)
		if err := setBasicType(reflect.ValueOf(defaults).Elem().Field(i), description, fieldName, defaul, field.Type.String()); err != nil {
			return err
		}
	}
	return nil
}

func setBasicType(value reflect.Value, desc string, name string, def reflect.Value, typ string) error {
	fmt.Printf("\n- %s\n%s(%s) [%v]: ", desc, name, typ, def)
	switch typ {
	case "string":
		value.SetString(askString(def.String()))
	case "int", "int64":
		value.SetInt(askInt64(def.Int()))
	case "bool":
		value.SetBool(askBool(def.Bool()))
	case "time.Duration":
		value.Set(reflect.ValueOf(askDuration(time.Duration(def.Int()))))
	case "*big.Int":
		value.Set(reflect.ValueOf(askBigInt()))
	default:
		return fmt.Errorf("unknown basic type <%s>", typ)
	}
	return nil
}

func askString(def string) string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()
	if input == "" {
		return def
	}
	return input
}

func askInt64(def int64) int64 {
	input := askString(fmt.Sprintf("%d", def))
	n, err := strconv.Atoi(input)
	if err != nil {
		return def
	}
	return int64(n)
}

func askBool(def bool) bool {
	n, err := strconv.ParseBool(askString(fmt.Sprintf("%v", def)))
	if err != nil {
		return def
	}
	return n
}

func askDuration(def time.Duration) time.Duration {
	t, err := time.ParseDuration(askString(fmt.Sprintf("%s", def)))
	if err != nil {
		return def
	}
	return t
}

func askBigInt() *big.Int {
	n := big.Int{}
	ret, _ := n.SetString(askString("0"), 10)
	return ret
}

func defaultConfiguration(tpl io.Reader) (*config.Configuration, error) {
	conf := config.Configuration{}
	in, err := io.ReadAll(tpl)
	if err != nil {
		return nil, err
	}
	if err := toml.Unmarshal(in, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func sub(s string) string {
	o := ""
	for i := 0; i < len(strings.Replace(s, "\n", "", -1)); i++ {
		o += "-"
	}
	return fmt.Sprintf("%s\r\n%s", s, o)
}
