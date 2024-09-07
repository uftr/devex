# Generate and Deploy Terraform modules

## Description

This repository contains a cli tool (vdex) that generates terraform main and deploys the desired terraform modules as per the developer needs.

The tool vdex (Vio Developer Experience) simplifies the development process for the terraform module users by performing below activities:
 - Generates the main.tf as per the input (or default) configuration
 - Deploys the terraform resources.

## How to Install

### Clone the repo and build locally

git clone https://github.com/uftr/vio.git

cd vio

To build, use either make or go build & go install
option 1: 
make

Option 2:
go build
go install

vdex binary will be generated.

If you don't have a permission to copy files to /usr/bin/local or GOPATH, then export the current path
Linux: export PATH=$PATH:.

Note: for windows the binary is vdex.exe instead of vdex

### Download the prebuilt binary

Download the binary from the repo https://github.com/uftr/vio/

### Dependency

terraform is required to use this tool. vdex has external dependency on terraform binary to execute plan and apply.
 **_NOTE_**: The package and binary have been tested on Windows and Linux environment with go 1.23 and terraform 1.95

## Usage

vdex init | plan [-s] | apply [-s]

### Help
Usage:
vdex init | plan [-s] | apply [-s]
    init       - takes user input for REPLACE-ME values and stores the config in sys/<SYSTEM-NAME>/
                 <SYSTEM-NAME> is one of the user input
    plan  [-s] - generates the main.tf file with the user values in sys/<SYSTEM-NAME>/.cache and
                 executes terraform init & plan        
                 if -s option is passed, terraform init will be skipped
    apply [-s] - similar plan but terraform apply is executed instead of terraform plan
                 if -s option is passed, terraform init will be skipped
    help       - this usage text

Below commands displays the usage help text
vdex
vdex -help
vdex --help
vdex ?

### vdex init

Interactively promts the user to provide vaules of the REPLACE-ME variables that are present in the module template main.tf, and stored the user input in a configuration file under <src/<SYSTEM-NAME/config.txt>
User is shown the default values of the variables. Present <Enter> to keep the default value unchanged.
Once the input is given, press <Enter> to proceed further.
If a variable doesn't have a default value and user skips by pressing <Enter>, then user is re-promted 3 times for the input.

> **_NOTE_**: main.tf by deault is expected in the working directory of the user from where vdex is invoked.

init will create src/ folder in the current workspace if it doesn't exist.

User can directly edit the generated configuration file <src/<SYSTEM-NAME>/config.txt> and replace the values for the variables. The configuration files accepts standard terraform comment (# or //) at the start of the any line.

Format of the configuration file:

Configuration data is stored as series of (key = value) pairs, each pair in a new line.

example, consider the below module template

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

The generated configuration looks like below:

module "echo".foo = "Platform"
provider "aws".default_tags.tags."System-Name" = "test"
environment = default

> **_NOTE_**: environment variable supports multiple environments and user is prompted to enter the desired environment. It has default environment and it is optional for user to change it.

> **_NOTE_**: User can edit the right hand values but avoid changing key names in the configuration file, unless change aligns with the module template.

>vdex init can process all valid terraform files with multiple level of hierarchy.

### vdex plan

Reads the configuration file and generate the main.tf file, which calls the Terraform module and configures the backend. The generated file will be stored in the <src/<systems-name>/.cache/main.tf> folder.
Subsequently, it runs the terraform init and terraform plan on the generated folder. 

If -s option is specified, terraform init is skipped.

### vdex apply

Reads the configuration file and generate the main.tf file, which calls the Terraform module and configures the backend. The generated file will be stored in the <src/<systems-name>/.cache/main.tf> folder.
Subsequently, it runs the Terraform init and Terraform apply on the generated folder.

If -s option is specified, terraform init is skipped.

## Special Features

### Multiple Environments

The configuration supports environment variable to support multiple environments. During init, the user is prompted to enter the desired environment. It is set to default environment and it is optional for user to change it. User can skip setting the value.

To maintain multiple environments, run vdex init and change value of this variable or directly edit the configuration file.

For instances environment can be set to dev, stage, prod etc. The vdex plan and apply commands, creates/maintains the underlying workspaces as per the configured environment.

If environment is set to prod, then vdex takes care of creating(if does not exist) and switching to the 'prod' workspace. If the 'prod' workspace is already present, then vdex just switches to the workspace.

Multiple Configuration Files
> **_NOTE_**: Currently, same configuration file present in the system-name folder gets updated by init. In next version, multiple configuration files will be created, distinct for each environment.

### Multiple Systems

When user configures <system-name> value during init, this value is used by plan and apply to create the system-name folder under "sys/".

If user gives system-name as "ci", folder "sys/ci" gets created. In next init run, if user gives system-name as "cd", then  folder "sys/cd" gets created.

When plan and apply are executed, all system folders under "sys/" gets processed. In this scenario, both the config files "sys/ci/config.txt" & "sys/cd/config.txt" gets processed.
