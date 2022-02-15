Data structure:

Primary Key:
    type - 2 bytes
    primary key   - 8 bytes

Value - Proto OR Avro

Index, value is empty:
    type - 2 bytes
    index - variable length
    primary key - 8 bytes

Foreign Key, value is empty:
    type - 2 bytes
    parent key - 8 bytes
    child key - 8 bytes

Sequence Key (Sequence Maintained by Badger)
    type - 2 bytes
    id   - 2 bytes

Item.UserMeta
    Avro - 1
    Protobuf - 2
    JSON - 3
    
Registry stores type information. All keys and values must be registered.
Types IDs 1-1024 are reserved

Stored copy. Used by the registry. Initialization sequence:

1. The registry registers internal types
2. The app registers app types
3. The registry loads stored copy
4. The registry compares stored copy against memory copy
5. The registry updates the stored copy or returns an error
6. The registry is ready to be used

