# sh-id-platform




### How to run the server locally?
In order to put to run the server you should follow the next steps:

1. Start docker containers:
```
make up
```

2. Create the database and run migrations:
```
make db/migrate
```

3. Set up vault token:
Add or modify the key store configuration in the config.toml file:
```toml
[KeyStore]
    Address="http://localhost:8200"
    # In testing mode this value should be taken from ./infrastructure/local/.vault/data/init.out
    Token="hvs.YxU2dLZljGpqLyPYu6VeYJde" 
    PluginIden3MountPath="iden3"
```


### Third party tools
If you want to execute the github actions locally visit https://github.com/nektos/act