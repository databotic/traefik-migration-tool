
# Traefik Migration Tool

A tool to migrate from Traefik v2 to Traefik v3.

## Description

This tool helps you convert Traefik v2 IngressRoute and Middleware configurations to Traefik v3. It can handle both YAML conversion and direct migration of Kubernetes resources from `traefik.containo.us/v1alpha1` to `traefik.io/v1alpha1`.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)

## Installation

To install the Traefik Migration Tool, download the binary from the release page.

## Usage

The tool offers several commands to help with the migration process:

```sh
A tool to migrate from Traefik v2 to Traefik v3.

Usage:
  traefik-migration-tool [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  convert     Convert Traefik v2 kubernetes resources to v3
  help        Help about any command
  migrate     Migrate existing Traefik v2 kubernetes resources to v3
  version     Display version

Flags:
  -h, --help   help for traefik-migration-tool

Use "traefik-migration-tool [command] --help" for more information about a command.
```

### `convert`

Convert Traefik v2 Kubernetes resources to v3.

```sh
traefik-migration-tool convert -f path/to/your/v2-ingressroute.yaml
```

### `migrate`

Migrate existing Traefik v2 Kubernetes resources to v3.

```sh
traefik-migration-tool migrate -h
```
