# Configuration File

## maskPRTC

```{ .yaml .annotate }
targetPRTC: name
```

## maskClock

```{ .yaml .annotate }
targetClock: name
```

## clockTargets

Multiple roles can be specificed for post-processing through `clockTargets` in the config file.

```{ .yaml .annotate }
clockTargets:
  - name: "gnss"
  - name: "tgm"
  - name: "tbc"
  - name: "toc"
```

## transient

## steady
