# Generate and Deploy Terraform modules

 **_NOTE_**: The package and binary have been tested only on Windows environment

## Overview

This repository contains a cli tool (vdex) that generates terraform main and deploys the desired terraform modules as per the developer needs.

## Purpose

The tool vdex (Vio Developer Experience) simplifies the development process for the terraform module users by performing below activities:
 - Generates the main.tf as per the input (or default) configuration
 - Deploys the terraform resources.

## How to Build

### Clone the repo and build locally

git clone https://github.com/uftr/vio

cd vio

go build

vdex binary will be generated.
export the image path to PATH environmental variable
Note: for windows the binary is vdex.exe instead of vdex

### Download the prebuilt binary

Download the binary from the repo https://github.com/uftr/vio

## Usage

vdex init | plan [-s] | apply [-s]

-s option is an optional argument in plan and apply

Note: for windows the binary is vdex.exe instead of vdex
Dependency: vdex has external dependency on terraform binary to execute plan and apply.

### Help
Below commands displays the usage help text
vdex
vdex -help
vdex ?

### vdex init

Interactively promts the user to provide vaules of the REPLACE-ME variables that are present in the module template main.tf, and stored the user input in a configuration file under <src/<SYSTEM-NAME/config.txt>
User is shown the default values of the variables. Present <Enter> to keep the default value unchanged.
Once the input is given, press <Enter> to proceed further.
If a variable doesn't have a default value and user skips by pressing <Enter>, then user is re-promted 3 times for the input.

> **_NOTE_**: main.tf by deault is expected in the working directory of the user.

User can directly edit the generated configuration file <src/<SYSTEM-NAME/config.txt> and replace the values for the variables. The configuration files accepts standard terraform comment (# or //) at the start of the any line.

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

> **_NOTE_**: User can edit the right hand values but avoid changing key names in the configuration file, unless change aligns with the module template.

> **_NOTE_**: vdex init can process all valid terraform files with multiple level of hierarchy.

### vdex plan

Reads the configuration file and generate the main.tf file, which calls the Terraform module and configures the backend. The generated file will be stored in the src/<systems-name>/.cache folder.
Then, runs the Terraform init and Terraform plan on the generated folder. 

If -s option is specified, terraform init is skipped.

### vdex apply

Reads the configuration file and generate the main.tf file, which calls the Terraform module and configures the backend. The generated file will be stored in the src/<systems-name>/.cache folder.
Then, runs the Terraform init and Terraform apply on the generated folder.

If -s option is specified, terraform init is skipped.