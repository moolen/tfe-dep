terraform {
  cloud {
    organization = "moolen"

    workspaces {
      name = "tdep-test"
    }
  }
}

data "tfe_outputs" "a" {
  organization = "moolen"
  workspace = "tdep-a"
}

output "hello" {
  sensitive = true
  value = data.tfe_outputs.a.values.hello
}