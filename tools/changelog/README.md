# Changelog creator

This tool creates a changelog from a list of PRs following the [conventional commit](https://www.conventionalcommits.org/en/v1.0.0/) format.

Steps to execute the tool:
1. Copy sample input file into input.md
 ```
cp input-sample.md input.md
```
2. Edit `input.md` and add the PRs you want to include in the changelog (generate release notes)
3. Execute the program

   
With Go:
```
go run main.go
```

With Binaries:
```
./bin/main
```

4. Copy the output from `changelog.md` into the changelog file


Build the project (go installed required):
```
go build -o bin/changelog main.go
```