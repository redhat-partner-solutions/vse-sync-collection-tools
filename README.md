# synchronization-testsuites
The purpose of this project is to make a single repository for all the existent used synchronization testing suites for the OpenShift platform, typically testing suites that target  different levels of testing (integration, operator testing, performance testing)

## Tree structure

```
.
├── README.md
├── doc
│   └── LIST_OF_TESTS.md
└── tests
    ├── common
    │   ├── cnf-features-deploy
    │   └── ptp-operator
    ├── extra
    ├── tbc
    ├── tgm
    │   ├── featurecalnex
    │   │   └── calnex.go
    │   └── ptpsynce_test.go
    └── tsc
```

## Installation Instructions

For repo to allow you to make changes to the specific test suite
Git clone repo

```console
git clone --recursive https://github.com/redhat-partner-solutions/synchronization-testsuites.git
```

## Update Instructions

### Update a single test suite

The following example uses openshift-ptp:

```
cd tests/common/openshift/ptp-operator
git pull
```

### Update all the synchronization test suites

```
cd synchronization-testsuites
git submodule foreach git pull
```

## Contributing to the repo

To contribute to the main repo acting as wrapper, send a pull request.

### Add a Testing Suite

How to add a new testing suite

### Update Existent Testing Suite

Create a PR to te existing testing suite

### Current Synchronization Testing Suites

Location: [/tests/common](/common)

Common testing suites: 

[ptp operator conformance testing](https://github.com/openshift/ptp-operator.git) - ptp operator testing

[cnf ptp tests](https://github.com/openshift-kni/cnf-features-deploy.git) - cnf ptp tests


Location: [/tests/tgm](/tgm) - "tests specific for T-GM role"

Location: [/tests/tbc](/tbc) - "tests specific for T-BC role"

Location: [/tests/tsc](/tsc) - "tests specific for T-SC role"

Location: [/tests/extra](/extra) - "extra sync tests suites"
