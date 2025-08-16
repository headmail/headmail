# Headmail Helm Chart

This chart deploys:
- headmail backend exposing public web (/) and admin API (/api)
- optional frontend static server that serves admin web at '/'

## Usage examples:

**Install with default values:**
```bash
helm upgrade --install headmail chart/headmail
```

**Override backend image tag:**
```bash
helm upgrade --install headmail chart/headmail --set backend.image.tag=0.0.1
```
