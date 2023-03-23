# Multichain bridge watcher

This project is responsible to watch the multichain bridge and track if the bridging works
correctly from XRPL to Coreum.

## Build

```bash
go build -o multichain-auditor
chmod +x multichain-auditor
```

## Use

### Help

```bash
./multichain-auditor help
```

### Export courem incoming transactions

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
./multichain-auditor discrepancy export --from-date-time="2023-03-23 00:00:00" --to-date-time="2023-01-01 00:00:00"
```

Pay attention, when you use the time boundaries, the discrepancies will be found within that range. Hence,
If the tx on the cream side was created after the `from-data-time`, the discrepancy will be found. It means in most cases better to 
use `to-date-time` only, where the `to-date-time` is the end of the fully paid period.
