# Mermaid

The following mermaid diagram renders as box-drawing art:

```mermaid
flowchart LR
    A[Build] --> B{Tests pass?}
    B -->|yes| C[Release]
    B -->|no| D[Fix bugs]
    D --> C
```
