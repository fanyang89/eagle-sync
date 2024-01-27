# eaglexport

Export local eagle library with smart folder filter.

## Usage

```bash
# export to local fs
eaglexport export --library <your-eagle-library> --dst <output-dir>

# export to smb
eaglexport export --library <your-eagle-library> --dst smb://127.0.0.1/share0/tmp --smb-user foo --smb-password bar
```

## License

MIT
