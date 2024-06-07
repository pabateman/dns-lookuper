# DNS Lookuper

DNS Lookuper is simple utility that resolves your list of domain names into their addresses. Supports several output formats such as `yaml`, `json`, `csv` and custom template as well.

## Installation

To be described later.

## Usage

Lookuper works with command line args as well as a config file.

The utility uses the concept of tasks. You can configure only one task through the command line, whereas through a configuration file, you can configure multiple tasks.

### Command line

Example with command line:

```bash
$ dns-lookuper -f ./testdata/lists/1.lst -f ./testdata/lists/2.lst -o - -m all -r hosts
```

Console output:

```hosts
cloudflare.com 104.16.133.229
cloudflare.com 104.16.132.229
cloudflare.com 2606:4700::6810:84e5
cloudflare.com 2606:4700::6810:85e5
google.com 173.194.220.138
google.com 173.194.220.102
google.com 173.194.220.100
google.com 173.194.220.101
google.com 173.194.220.113
google.com 173.194.220.139
google.com 2a00:1450:4010:c0e::8b
google.com 2a00:1450:4010:c0e::66
google.com 2a00:1450:4010:c0e::71
google.com 2a00:1450:4010:c0e::64
hashicorp.com 76.76.21.21
linked.in 108.174.10.24
linked.in 2620:109:c002::6cae:a18
releases.hashicorp.com 108.157.229.119
rpm.releases.hashicorp.com 3.164.230.2
rpm.releases.hashicorp.com 3.164.230.117
rpm.releases.hashicorp.com 3.164.230.48
rpm.releases.hashicorp.com 3.164.230.56
rpm.releases.hashicorp.com 2600:9000:25f7:d200:18:566b:ecc0:93a1
rpm.releases.hashicorp.com 2600:9000:25f7:a400:18:566b:ecc0:93a1
rpm.releases.hashicorp.com 2600:9000:25f7:ca00:18:566b:ecc0:93a1
rpm.releases.hashicorp.com 2600:9000:25f7:8c00:18:566b:ecc0:93a1
rpm.releases.hashicorp.com 2600:9000:25f7:0:18:566b:ecc0:93a1
rpm.releases.hashicorp.com 2600:9000:25f7:6e00:18:566b:ecc0:93a1
rpm.releases.hashicorp.com 2600:9000:25f7:9000:18:566b:ecc0:93a1
rpm.releases.hashicorp.com 2600:9000:25f7:8200:18:566b:ecc0:93a1
terraform.io 76.76.21.21
```

### Configuration file

First, you need a config file:

```yaml
tasks:
  - files:
      - ./../lists/1.lst
      - ./../lists/2.lst
    output: "-"
    format: list
    mode: ipv4
  - files:
      - ./../lists/1.lst
      - ./../lists/2.lst
    output: result_ipv6.txt
    format: list
    mode: ipv6
```

This example stored in [testdata/configs/many.yaml](/testdata/configs/many.yaml).

```bash
$ dns-lookuper -c testdata/configs/many.yaml
```

The result of the execution will be console output:

```list
104.16.132.229
104.16.133.229
108.157.229.119
108.174.10.24
173.194.222.100
173.194.222.101
173.194.222.102
173.194.222.113
173.194.222.138
173.194.222.139
18.165.140.122
18.165.140.50
18.165.140.52
18.165.140.56
76.76.21.21
```

Additionally, a file will be stored in testdata/output/result_ipv6.txt with the following content:

```list
2600:9000:272c:3000:18:566b:ecc0:93a1
2600:9000:272c:4400:18:566b:ecc0:93a1
2600:9000:272c:4600:18:566b:ecc0:93a1
2600:9000:272c:4c00:18:566b:ecc0:93a1
2600:9000:272c:d200:18:566b:ecc0:93a1
2600:9000:272c:d600:18:566b:ecc0:93a1
2600:9000:272c:da00:18:566b:ecc0:93a1
2600:9000:272c:fc00:18:566b:ecc0:93a1
2606:4700::6810:84e5
2606:4700::6810:85e5
2620:109:c002::6cae:a18
2a00:1450:4010:c03::64
2a00:1450:4010:c03::65
2a00:1450:4010:c03::71
2a00:1450:4010:c03::8b
```

**Important notice:** with the configuration file, only one task is allowed to print to the console (`/dev/stdout` or `/dev/stderr`) by design purposes.

### Daemon mode

DNS Lookuper supports a daemon mode, in which the utility executes continuously at a specified interval (1 minute by default). The interval must be specified in Go duration format, e.g., 30s, 5m, 3h, 1d, 5y. Similar to oneshot mode, there is support for command line options or a configuration file.

**Important notice:** Daemon mode not supports console output (`/dev/stdout` or `/dev/stderr`) by design purposes.

Example with command line:

```bash
$ dns-lookuper -f ./testdata/lists/1.lst -f ./testdata/lists/2.lst -o testdata/output/result.txt -m all -r hosts -d -i 10s
```

As result of the execution a file will be stored in testdata/output/result.txt and it will be updated every 10 seconds.

Example with config file:

Config file:

```yaml
settings:
  lookupTimeout: 2
  daemon:
    enabled: true
    interval: 30s
tasks:
  - files:
      - ../lists/1.lst
      - ../lists/2.lst
    output: ../output/daemonconfig.txt
    format: yaml

```

Stored in [testdata/configs/simple-daemon.yaml](/testdata/configs/simple-daemon.yaml).

Try it:

```bash
$ dns-lookuper -c testdata/configs/simple-daemon.yaml
```

As result of the execution a file will be stored in testdata/output/daemonconfig.txt and it will be updated every 30 seconds.

## Output formats

DNS Lookuper supports several output formats, including:

- Simple list
- Hosts file
- JSON
- YAML
- CSV
- Template

### Simple list

Just simple list of IP addresses, example:

```bash
$ dns-lookuper -f testdata/lists/1.lst -r list
```

Output:

```list
104.16.132.229
104.16.133.229
2606:4700::6810:84e5
2606:4700::6810:85e5
76.76.21.21
```

### Hosts file

Classic `/etc/hosts` format, example:

```bash
$ dns-lookuper -f testdata/lists/1.lst -r hosts
```

Output:

```hosts
104.16.133.229 cloudflare.com
104.16.132.229 cloudflare.com
2606:4700::6810:84e5 cloudflare.com
2606:4700::6810:85e5 cloudflare.com
76.76.21.21 hashicorp.com
76.76.21.21 terraform.io
```

### JSON

A list of objects with the name as a string and adresses as a list:

```bash
$ dns-lookuper -f testdata/lists/1.lst -r json
```

Output:

```json
[
  {
    "name": "cloudflare.com",
    "addresses": [
      "104.16.132.229",
      "104.16.133.229",
      "2606:4700::6810:84e5",
      "2606:4700::6810:85e5"
    ]
  },
  {
    "name": "hashicorp.com",
    "addresses": [
      "76.76.21.21"
    ]
  },
  {
    "name": "terraform.io",
    "addresses": [
      "76.76.21.21"
    ]
  }
```

### YAML

Similar to JSON, but YAML:

```bash
$ dns-lookuper -f testdata/lists/1.lst -r json
```

Output:

```yaml
- addresses:
  - 104.16.133.229
  - 104.16.132.229
  - 2606:4700::6810:84e5
  - 2606:4700::6810:85e5
  name: cloudflare.com
- addresses:
  - 76.76.21.21
  name: hashicorp.com
- addresses:
  - 76.76.21.21
  name: terraform.io
```

### CSV

A simple CSV file:

```bash
$ dns-lookuper -f testdata/lists/1.lst -r csv
```

Output:

```csv
name,address
cloudflare.com,104.16.132.229
cloudflare.com,104.16.133.229
cloudflare.com,2606:4700::6810:85e5
cloudflare.com,2606:4700::6810:84e5
hashicorp.com,76.76.21.21
terraform.io,76.76.21.21
```

### Template

Additionally, you can specify your own template for the lookup result for every task separately. The only available variables are `{{host}}` for the host and `{{address}}` for addresses:

```bash
$ dns-lookuper -f testdata/lists/1.lst -r template -t "there is {{host}} with address {{address}}"
```

Output:

```text
there is cloudflare.com with address 104.16.132.229
there is cloudflare.com with address 104.16.133.229
there is cloudflare.com with address 2606:4700::6810:85e5
there is cloudflare.com with address 2606:4700::6810:84e5
there is hashicorp.com with address 76.76.21.21
there is terraform.io with address 76.76.21.21
```
