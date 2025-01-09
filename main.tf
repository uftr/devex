module "echo" {
    source = "../"
    foo = 5 // REPLACE-ME
    bar = "hello" // REPLACE-ME
    items = ["30","40"] // REPLACE-ME
}

# test1
# test2
provider "aws" {
    region = "us-east-1"
    default_tags {   
        tags = {
            "Team"        = "REPLACE-ME"
            "System-Name" = "REPLACE-ME"
        } 
    }
}
/*
 * comment
 */

terraform {
    backend "local" {}
}