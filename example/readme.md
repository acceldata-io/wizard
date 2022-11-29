# Example Program

An example program that implements the wizard both with chan and without chan methods.

```shell
├── server
│   ├── example_server
│   ├── go.mod
│   └── main.go
└── tasks
    ├── files
    │   ├── example.json
    │   └── example.service.tmpl
    ├── go.mod
    ├── go.sum
    └── main.go
```

- The server folder has a basic http server to run as a systemd service by the example program.
- The tasks folder has the example.json file under `tasks/files/example.json`. This file is the input config for the wizard. It follows the Wizard DSL and performs the following tasks:
  - Create user
  - Create a dir
  - Copy example_server to the above created dir
  - Create the systemd service file from the template file
  - Start the systemd service
- The systemd service file step creates a register. This register is evaluated in when condition for the next step to start the systemd service.
  
>All the above files and services will be owned by the user created in the first action.

---

## Compile the example_server for Linux

```shell
cd ./server/
```

```shell
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -installsuffix cgo '-ldflags=-w -s' .
```

This will create the `example_server` binary

---

## Compile the example binary for Linux

```shell
cd ./tasks/
```

```shell
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -installsuffix cgo '-ldflags=-w -s' .
```

This will create the `example` binary

---

## Run the example program

The example program can run both the implementations of wizard. The program needs to be ran as a sudo user with the desired command line argument which specifies the type of implementation.

- Run without logs as chan

```shell
sudo example non-chan
```

- Run with logs as chan

```shell
sudo example chan
```

The output logs in both the cases are printed to the stdout.

---
