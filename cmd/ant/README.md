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

- new project

```sh
NAME:
   ant new - Create a new ant project

USAGE:
   ant new [options] [arguments...]
 or
   ant new {app_path} [options except -app_path] [arguments...]

OPTIONS:
   --app_path value, -a value  Specifies the path(relative/absolute) of the project
```

- run project

```sh
NAME:
   ant run - Compile and run gracefully (monitor changes) an any existing go project

USAGE:
   ant run [options] [arguments...]
 or
   ant run [options except -app_path] [arguments...] {app_path}

OPTIONS:
   --watch_exts value, -x value  Specified to increase the listening file suffix (default: ".go", ".ini", ".yaml", ".toml", ".xml")
   --app_path value, -a value    Specifies the path(relative/absolute) of the project
```

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