# Test Execution

## Common Example Commands

### Run all the tests with verbose option:

```
ginkgo -v run tests -- --ginkgo.focus-file=tbc.go --config=../razorbill_test_config.yaml --env=../razorbill_env.yaml
```

### Run the Telecom GrandMaster and GNSS tests:

```
ginkgo run tests -- --ginkgo.focus-file=tbc.go --config=../razorbill_test_config.yaml --env=../razorbill_env.yaml
```

### Run the Telecom Boundary Clock tests wih verbose option:

```
ginkgo -v run tests -- --ginkgo.focus-file=tbc.go --config=../razorbill_test_config.yaml --env=../razorbill_env.yaml
```
