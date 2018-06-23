# IAM Role Annotator
This Kubernetes controller is watching Deployment objects using the Kubernetes API. Whenever a Deployment is created or updated,
 the controller will check if the Deployment contains the `armesto.net/iam-role-annotator` annotation, and, in that case, add the `iam.amazonaws.com/role` annotation containing the appropiate IAM Role.

[![Build Status](https://travis-ci.org/fiunchinho/iam-role-annotator.svg?branch=master)](https://travis-ci.org/fiunchinho/iam-role-annotator)

The IAM Role Annotator assumes that an IAM Role is already created for every application. The IAM Role ARN used in the annotation
 will be an ARN of the form `arn:aws:iam::<AWS_ACCOUNT_ID>:role/<APPLICATION_NAME>`, where the application name is the name of the `Deployment` object.

## Build
We provide a Makefile that you can use to build this application
```bash
$ make
```

Or if you are running linux
```bash
$ make build/iam-role-annotator-linux-amd64
```

## Tests
You can run the tests using the Makefile
```bash
$ make test
```

## Usage
You can start the application with the following command
```bash
$ go run ./cmd/* --namespace your-namespace --aws-account-id 12345
```

### Parameters
These are the available parameters
- **namespace**: Only Deployments in this namespace will be watched
- **aws-account-id**: The AWS account id used in the role's ARN
- **resync-seconds**: The controller will reprocess all watched objects every `resync-seconds` seconds
- **kubeconfig**: Kubernetes configuration file used to connect to the cluster, only used when running the controller outside the cluster

## Releasing
This application is package using Docker containers that are published [in this repository]()
```bash
$ make release
```
