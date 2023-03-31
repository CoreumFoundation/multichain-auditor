# Multichain bridge watcher

This project is responsible for watching the multichain bridge and track if the bridging works
correctly from XRPL to Coreum.

## Build

```bash
go build -o multichain-auditor
```

## Use

### Help

```bash
./multichain-auditor help
```

### Export coreum incoming transactions

```bash
./multichain-auditor coreum export-incoming
```

### Export coreum outgoing transactions

```bash
./multichain-auditor coreum export-outgoing 
```

### Export xrpl incoming transactions

```bash
./multichain-auditor xrpl export-incoming
```

### Export discrepancies

```bash
./multichain-auditor discrepancy export
```

### Export discrepancies and include rows even if there are no discrepancies

```bash
./multichain-auditor discrepancy export --include-all=true
```

### Export discrepancies with time boundaries

```bash
./multichain-auditor discrepancy export --before-date-time="2023-03-23 00:00:00" --after-date-time="2023-01-01 00:00:00"
```

### Rescan orphan tx discrepancies with multichain

```bash
./multichain-auditor discrepancy rescan
```

### Print summary print

```bash
./multichain-auditor summary print
```


