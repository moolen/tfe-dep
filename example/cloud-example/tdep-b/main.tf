terraform {
  cloud {
    organization = "moolen"

    workspaces {
      name = "tdep-b"
    }
  }
}

output "hello" {
  value = "hello"
}

data "tfe_outputs" "test" {
  organization = "moolen"
  workspace = "tdep-test"
}

output "hello-from-test" {
  sensitive = true
  value = data.tfe_outputs.test.values.hello
}