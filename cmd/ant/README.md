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
   --template value, -t value    The template for code generation(relative/absolute)
   --app_path value, -p value  The path(relative/absolute) of the project
```

example: `ant gen -t ./__ant__tpl__.go -p ./myant` or default `ant gen myant`

template file `__ant__tpl__.go` demo:

```go
// package __ANT__TPL__ is the project template
package __ANT__TPL__

// __API__PULL__ register PULL router:
//  /home
//  /math/divide
type __API__PULL__ interface {
  Home(*struct{}) *HomeReply
  Math
}

// __API__PUSH__ register PUSH router:
//  /stat
type __API__PUSH__ interface {
  Stat(*StatArgs)
}

// Math controller
type Math interface {
  // Divide handler
  Divide(*DivideArgs) *DivideReply
}

// HomeReply home reply
type HomeReply struct {
  Content string // text
}

type (
  // DivideArgs divide api args
  DivideArgs struct {
    // dividend
    A float64
    // divisor
    B float64 `param:"<range: 0.01:100000>"`
  }
  // DivideReply divide api result
  DivideReply struct {
    // quotient
    C float64
  }
)

// StatArgs stat handler args
type StatArgs struct {
  Ts int64 // timestamps
}
```

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