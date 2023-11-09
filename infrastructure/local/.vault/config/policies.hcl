path "iden3/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "auth/userpass/users/issuernode" {
  capabilities = [ "update" ]
  allowed_parameters = {
   "password" = [ ]
  }
}