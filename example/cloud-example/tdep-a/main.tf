terraform {
  cloud {
    organization = "moolen"

    workspaces {
      name = "tdep-a"
    }
  }
}

data "tfe_outputs" "b" {
  organization = "moolen"
  workspace = "tdep-b"
}

output "hello" {
  sensitive = true
  value = data.tfe_outputs.b.values.hello
}