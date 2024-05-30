# DNS Lookuper

DNS Lookuper is simple utility that resolves your list of domain names into their addresses. Supports several output formats such as `yaml`, `json`, `csv` and custom template as well.

## Usage

Lookuper works with command line args as well as a config file.

The utility uses the concept of tasks. You can configure only one task through the command line, whereas through a configuration file, you can configure multiple tasks.

### Command line

Example with command line:

```bash
dns-lookuper -f ./testdata/lists/1.lst -f ./testdata/lists/2.lst -o - -m all -r hosts
```

Output:

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

First of all you need a config file:

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

### Daemon mode

## Output formats

### Simple list

### Hosts file

### JSON

### YAML

### CSV

### Template

## Distribution

TODO: Docker image and binary file
