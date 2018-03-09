# ant

Command ant is a deployment tools of [ant](https://github.com/henrylee2cn/ant) microservice frameware.

## 1. Feature

- Quickly create a ant project
- Run ant project with hot compilation

## 2. Install

```sh
go install
```

## 3. Usage

- generate project

```
NAME:
   ant gen - Generate an ant project

USAGE:
   ant gen [command options] [arguments...]

OPTIONS:
   --script value, -s value    The script for code generation(relative/absolute)
   --app_path value, -p value  The path(relative/absolute) of the project
```

example: `ant gen -s ./test.ant -p ./myant`

- run project

```
NAME:
   ant run - Compile and run gracefully (monitor changes) an any existing go project

USAGE:
   ant run [options] [arguments...]
 or
   ant run [options except -app_path] [arguments...] {app_path}

OPTIONS:
   --watch_exts value, -x value  Specified to increase the listening file suffix (default: ".go", ".ini", ".yaml", ".toml", ".xml")
   --app_path value, -p value    The path(relative/absolute) of the project
```

example: `ant run -x .yaml -p myant` or `ant run -x .yaml myant`

## 4. Project Structure

The template for using ant development project.

```
├── README.md
├── main.go
├── api
│   ├── handlers.gen.go
│   ├── handlers.go
│   ├── router.gen.go
│   └── router.go
├── logic
│   └── xxx.go
├── sdk
│   ├── rpc.gen.go
│   ├── rpc.gen_test.go
│   ├── rpc.go
│   └── rpc_test.go
└── types
    ├── types.gen.go
    └── types.go
```

Desc:

- add `.gen` suffix to the file name of the automatically generated file