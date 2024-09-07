# Generate and Deploy Terraform modules

## Description

This repository contains a cli tool (vdex) that generates terraform main and deploys the desired terraform modules as per the developer needs.

The tool vdex (Vio Developer Experience) simplifies the development process for the terraform module users by performing below activities:
 - Generates the main.tf as per the input (or default) configuration
 - Deploys the terraform resources.

## How to Install

### Clone the repo and build locally

```
git clone https://github.com/uftr/vio.git

cd vio
```

To build and install, use either make or (go build & go install)

- option 1: vdex binary gets copied to /usr/local/bin/

```
make
```

- Option 2: vdex binary gets copied to GOPATH if exists

```
go build
go install
```

vdex binary will be generated.

If you don't have a permission to copy files to `/usr/bin/local` or `GOPATH`, then export the current path

Linux:
```
export PATH=$PATH:.
```

Note: for windows the binary is vdex.exe instead of vdex

### Download the prebuilt binary

Download the binary from the repo https://github.com/uftr/vio/

### Dependency

terraform is required to use this tool. vdex has external dependency on terraform binary to execute plan and apply.
 **_NOTE_**: The package and binary have been tested on Windows and Linux environment with `go 1.2`3 and `terraform 1.95`

## Usage
```
vdex init [envName] | plan [-s] [envName] | apply [-s] [envName]

vdex init [envName] | [-s] plan [envName] | [-s] apply [envName]
```

### Help
Usage:
```
vdex init | plan [-s] | apply [-s]
```

```
-    init  [envName]
                - takes user input for REPLACE-ME values and stores the config in `sys/<SYSTEM-NAME>/
                 <SYSTEM-NAME>` is one of the user input.
                - envName is optional argument and if passed, it is treated as the environment which creates
                 a distrinct config file for the environment. It generated the file `<envName>-config.txt`

-    plan  [-s] [envName] 
                - generates the main.tf file with the user values in `sys/<SYSTEM-NAME>/.cache` and
                 executes terraform init & plan        
                - If -s option is passed, terraform init will be skipped
                - envName is optional argument and if passed, it is treated as the environment causing it to
                 process the config file named `<envName>-config.txt`. New workspace named `<envName>` will be used/setup by terraform init and plan.

-    apply [-s] [envName]
                - similar plan but terraform apply is executed instead of terraform plan
                - If -s option is passed, terraform init will be skipped.
                - envName is optional argument and if passed, it is treated as the environment causing it to
                 process the config file named `<envName>-config.txt`. New workspace named `<envName>` will be used/setup by terraform init and apply.

-    help       - this usage text
```

Below commands displays the usage help text

```
vdex
vdex -help
vdex --help
vdex ?
```

### vdex init

Interactively promts the user to provide vaules of the `REPLACE-ME` variables that are present in the module template main.tf, and stored the user input in a configuration file under `<src/<SYSTEM-NAME/config.txt>`
User is shown the default values of the variables. Present `<Enter>` to keep the default value unchanged.
Once the input is given, press `<Enter>` to proceed further.

If a variable doesn't have a default value and user skips by pressing `<Enter>`, then user is re-promted 3 times for the input.

> **_NOTE_**: main.tf by deault is expected in the working directory of the user from where vdex is invoked.

***init*** will create `src/` folder in the current workspace if it doesn't exist.

> [!CAUTION]
> User can directly edit the generated configuration file `<src/<SYSTEM-NAME>/config.txt>` and replace the values for the variables. The configuration files accepts standard terraform **_comment (# or //)_** at the start of the any line.

Format of the configuration file:

Configuration data is stored as series of **(key = value)** pairs, each pair in a new line.

example, consider the below module template:
```
module "echo" {
    foo = 5 // REPLACE-ME
}

provider "aws" {
    default_tags {   
        tags = {
            "System-Name" = "REPLACE-ME"
        } 
    }
}
```

The generated configuration looks like below:
```
module "echo".foo = "Platform"

provider "aws".default_tags.tags."System-Name" = "test"

environment = default
```

> **_NOTE_**: environment variable supports multiple environments and user is prompted to enter the desired environment. It has default environment and it is optional for user to change it.

> **_NOTE_**: User can edit the right hand values but avoid changing key names in the configuration file, unless change aligns with the module template.

>**vdex init** can process all valid terraform files with multiple level of hierarchy.

### vdex plan

Reads the configuration file and generate the main.tf file, which calls the Terraform module and configures the backend. The generated file will be stored in the `<src/<systems-name>/.cache/main.tf>`.
Subsequently, it runs the terraform init and terraform plan on the generated folder. 

If `-s` option is specified, ***terraform init*** is skipped.

### vdex apply

Reads the configuration file and generate the main.tf file, which calls the Terraform module and configures the backend. The generated file will be stored in the `<src/<systems-name>/.cache/main.tf>`.
Subsequently, it runs the Terraform init and Terraform apply on the generated folder.

If `-s` option is specified, ***terraform init*** is skipped.

## Special Features

### Multiple Environments

- Option 1: Multiple Configuration files - per environment

vdex cli supports optional argument to specify the environment in the cli argument itself.

```
vdex init [-s] [envName]
vdex plan [-s] [envName]
vdex apply [-s] [envName]
```

For example:
```
  "vdex init dev" results in creation of configuration file name "dev-config.txt".

  "vdex init prod" results in creation of configuration file name "prod-config.txt".

  vdex plan dev  - processes the dev configuration named "dev-config.txt".

  vdex plan prod - processes the dev configuration named "prod-config.txt".
```

The environment name is also stored as one of the config items in the configuration file.

The vdex **plan** and **apply** commands creates/maintains the underlying terraform workspaces as per the configured environment.

If environment is set to `prod`, then vdex takes care of creating(if does not exist) and switching to the `prod` workspace. If the `prod` workspace is already present, then vdex just switches to the workspace.

- Option 2: single config file - current environment

This option is slight variation of option 1. Here, user need not pass the environment in the cli argument, instead the environment can be set in the configuration file itself.

During init, the user is optionally prompted to enter the desired environment. Initially, it is set to default environment and it is optional for user to change it. User can skip setting the value.

To maintain multiple environments, directly edit environment variable in the configuration file or run vdex init and change value of this variable. In this model, same configuration file (`config.txt`) present in the system-name folder gets updated by init.

The vdex **plan and apply** looks at the environmental variable present in the configuration file (`config.txt`) and sets up (created/swith) the terraform workspaces accprdingly. 

### Multiple Systems

When user configures `<system-name>` value during init, this value is used by plan and apply to create the system-name folder under "sys/".

If user gives `system-name` as "ci", folder `"sys/ci"` gets created. In next init run, if user gives `system-name` as "cd", then  folder `"sys/cd"` gets created.

When plan and apply are executed, all system folders under "sys/" gets processed. In this scenario, both the config files `"sys/ci/config.txt"` & `"sys/cd/config.txt"` gets processed.
