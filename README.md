# BOT (Bucket Object Transferer)

![GitHub Release](https://img.shields.io/github/v/release/sebastocorp/bot)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/sebastocorp/bot)
![GitHub License](https://img.shields.io/github/license/sebastocorp/bot)

![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/sebastocorp/bot/total)
![GitHub forks](https://img.shields.io/github/forks/sebastocorp/bot)
![GitHub Repo stars](https://img.shields.io/github/stars/sebastocorp/bot)

![GitHub User's stars](https://img.shields.io/github/stars/sebastocorp)
![GitHub followers](https://img.shields.io/github/followers/sebastocorp)

## Description

BOT is a service dedicated to transferring objects from one object storage service to another and recording it in a MySQL database.

## Motivation

On many occasions, we require migrating objects from one bucket in an object storage to another, but performing this migration can be very costly if done immediately. Often, our intention is to carry out this migration gradually. BOT handles this transfer with each request received by its API.

TODO

## Flags

| Name                  | Command  | Default                              | Description |
|:---                   |:---      |:---                                  |:---         |
| `log-level`           | `server` | `info`                               | Verbosity level for logs |

## How to use

This project provides the binary files in differents architectures to make it easy to use wherever wanted.

### Configuration

Current configuration version: `v1alpha1`

## How does it work?

1. Checks the if object in backend bucket exist
2. Transfer the object from backend to frontend bucket

## Example

```sh
bot server
```

## How to collaborate

We are open to external collaborations for this project: improvements, bugfixes, whatever.

For doing it, open an issue to discuss the need of the changes, then:

- Fork the repository
- Make your changes to the code
- Open a PR and wait for review
