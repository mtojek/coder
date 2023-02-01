terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
      version = "0.6.6"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 2.22"
    }
  }
}

locals {
  username = data.coder_workspace.me.owner
}

data "coder_provisioner" "me" {
}

provider "docker" {
}

data "coder_workspace" "me" {
}

data "coder_parameter" "project_id" {
  name        = "Project ID"
  icon        = "/icon/azure.png"
  description = "This is the Project ID."
  default     = "12345"
  validation {
    regex = "^[a-z0-9]+$"
    error = "I'm sorry, my friend"
  }
}

data "coder_parameter" "random_string" {
  name        = "Random string"
  icon        = "/icon/aws.png"
  description = "This is a sample, random parameter."
  default     = "helloworld"
  mutable     = true
}

data "coder_parameter" "region" {
  name        = "Region"
  icon        = "/icon/database.svg"
  description = "These are options."
  mutable     = true
  option {
    name        = "US Central"
    description = "Select for central!"
    value       = "us-central1-a"
    icon        = "/icon/goland.svg"
  }
  option {
    name        = "US East"
    description = "Select for east!"
    value       = "us-east1-a"
    icon        = "/icon/folder.svg"
  }
  option {
    name        = "US West"
    description = "Select for west!"
    value       = "us-west2-a"
  }
}

data "coder_parameter" "production_env" {
  name        = "Production environment"
  icon        = "/icon/rider.svg"
  type        = "bool"
  description = "Is it production environment?"
  default     = "false"
  mutable     = true
}

data "coder_parameter" "compute_instances" {
  name        = "Compute instances"
  icon        = "/icon/rubymine.svg"
  type        = "number"
  description = "Let's set the expected number of instances."
  default     = "3"
  mutable     = true
  validation {
    min = 1
    max = 8
  }
}

resource "coder_agent" "main" {
  arch           = data.coder_provisioner.me.arch
  os             = "linux"
  startup_script = <<EOF
    #!/bin/sh
    # install and start code-server
    curl -fsSL https://code-server.dev/install.sh | sh -s -- --version 4.8.3
    code-server --auth none --port 13337
    EOF

  # These environment variables allow you to make Git commits right away after creating a
  # workspace. Note that they take precedence over configuration defined in ~/.gitconfig!
  # You can remove this block if you'd prefer to configure Git manually or using
  # dotfiles. (see docs/dotfiles.md)
  env = {
    GIT_AUTHOR_NAME     = "${data.coder_workspace.me.owner}"
    GIT_COMMITTER_NAME  = "${data.coder_workspace.me.owner}"
    GIT_AUTHOR_EMAIL    = "${data.coder_workspace.me.owner_email}"
    GIT_COMMITTER_EMAIL = "${data.coder_workspace.me.owner_email}"
  }
}

resource "coder_app" "code-server" {
  agent_id     = coder_agent.main.id
  slug         = "code-server"
  display_name = "code-server"
  url          = "http://localhost:13337/?folder=/home/${local.username}"
  icon         = "/icon/code.svg"
  subdomain    = false
  share        = "owner"

  healthcheck {
    url       = "http://localhost:13337/healthz"
    interval  = 5
    threshold = 6
  }
}


resource "docker_volume" "home_volume" {
  name = "coder-${data.coder_workspace.me.id}-home"
  # Protect the volume from being deleted due to changes in attributes.
  lifecycle {
    ignore_changes = all
  }
  # Add labels in Docker to keep track of orphan resources.
  labels {
    label = "coder.owner"
    value = data.coder_workspace.me.owner
  }
  labels {
    label = "coder.owner_id"
    value = data.coder_workspace.me.owner_id
  }
  labels {
    label = "coder.workspace_id"
    value = data.coder_workspace.me.id
  }
  # This field becomes outdated if the workspace is renamed but can
  # be useful for debugging or cleaning out dangling volumes.
  labels {
    label = "coder.workspace_name_at_creation"
    value = data.coder_workspace.me.name
  }
}


resource "docker_image" "main" {
  name = "coder-${data.coder_workspace.me.id}"
  build {
    path = "./build"
    build_args = {
      USER = local.username
    }
  }
  triggers = {
    dir_sha1 = sha1(join("", [for f in fileset(path.module, "build/*") : filesha1(f)]))
  }
}

resource "docker_container" "workspace" {
  count = data.coder_workspace.me.start_count
  image = docker_image.main.name
  # Uses lower() to avoid Docker restriction on container names.
  name = "coder-${data.coder_workspace.me.owner}-${lower(data.coder_workspace.me.name)}"
  # Hostname makes the shell more user friendly: coder@my-workspace:~$
  hostname = data.coder_workspace.me.name
  # Use the docker gateway if the access URL is 127.0.0.1
  entrypoint = ["sh", "-c", replace(coder_agent.main.init_script, "/localhost|127\\.0\\.0\\.1/", "host.docker.internal")]
  env = [
    "CODER_AGENT_TOKEN=${coder_agent.main.token}",
    "FOOBAR_PROJECT_ID=${data.coder_parameter.project_id.value}",
    "FOOBAR_RANDOM_STRING=${data.coder_parameter.random_string.value}",
    "FOOBAR_REGION=${data.coder_parameter.region.value}",
    "FOOBAR_PRODUCTION_ENV=${data.coder_parameter.production_env.value}",
    "FOOBAR_COMPUTE_INSTANCES=${data.coder_parameter.compute_instances.value}",
  ]
  host {
    host = "host.docker.internal"
    ip   = "host-gateway"
  }
  volumes {
    container_path = "/home/${local.username}"
    volume_name    = docker_volume.home_volume.name
    read_only      = false
  }
  # Add labels in Docker to keep track of orphan resources.
  labels {
    label = "coder.owner"
    value = data.coder_workspace.me.owner
  }
  labels {
    label = "coder.owner_id"
    value = data.coder_workspace.me.owner_id
  }
  labels {
    label = "coder.workspace_id"
    value = data.coder_workspace.me.id
  }
  labels {
    label = "coder.workspace_name"
    value = data.coder_workspace.me.name
  }
}
