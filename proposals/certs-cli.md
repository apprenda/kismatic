# Certificates CLI

Status: Proposal

Certificates are a requirement for running secure Kubernetes clusters. In a cluster bootstrapped by kismatic, all component endpoints are secured using TLS. Furthermore, certificates are also used for user authentication when there is no other authentication mechanism configured.

In order to facilitate the generation and management of certificates, a new command in the kismatic CLI is proposed.

## Use Cases
* As an operator, I want to grant cluster access to a new user by generating a client certificate
* As an operator, I want to grant cluster access to an external system by generating a client certificate

# Design

The following CLI commands are proposed for introduction to the `kismatic` binary:

## Certificate Generation
A generic, "swiss-army" style command is proposed for generating certificates:

```
kismatic certificates generate <name> [options]
```

Output: Generated key and certificate is placed in the generated directory. Key's filename is `<name>-key.pem`, and certificate's filename is `<name>.pem`.

Pre-conditions:
* CA is in the generated directory

Options:
* `--common-name`: Override the common name. If blank, use `<name>`.
* `--validity-period`: Specify the number of days this certificate should be valid for. Expiration date will be calculated relative to the machine's clock. If not specified, the validity period will be 365 days.
* `--subj-alt-names`: Comma-separated list of names that should be included in the certificate's subject alternative names field.
* `--organizations`: Comma-separated list of names that should be included in the certificate's organization field.
* `--overwrite`: Overwrite existing certificate if it already exists in the target directory.

Validity Period:
* Kismatic will print a warning if the chosen validity period is longer than the recommended 825 days (https://cabforum.org/2017/03/17/ballot-193-825-day-certificate-lifetimes/)

# Other considerations
* Ensure that our certificate validation code allows for custom SANs and organizations. I belive this
is already the case, but we should verify.