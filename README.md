# PlugMeter

Plugmeter is a simple daemon to monitoring energy consumption from Shelly Plugs. 

Features: 
* power monitoring and logging (in csv files)
* REST API
* Web UI
* optional plugs automatic detection

## Usage
### Configuration

Configuration can be done in the three following ways, each approach taking precedence over the one following it :

* Command line
* Environnement Variables
* Configuration File

### Command line

Use `plugmeter --help` for a list of all available flags.

### Environment Variables

Supported environnement variables, whose names loosely matche the command line flags: `UI_PORT`, `PLUG_DISCOVERY`, `PLUG_IPS`, `POLL_PERIOD`, `MAX_ERROR`, `LOG_LEVEL`, `CSV_OUT`, `CSV_FILE` and `DB_FILE`.

### Configuration file

The configuration file is a TOML-formated file and by default searched for in `/etc/plugmeter/` `$HOME/.plugmeter` and in the working directory. You can also set a full path to a config file using the `--conf <PATH_TO_CONFIG_FILE>` CLI flag.

See `plugmeter_conf.toml` for a full list of supported configuration options.

## Using as a docker container


`docker build -t plugmeter . `

 
```
docker run -d -p 3000:3000 -v ./out/:/out plugmeter
```


```
docker run -d  --env-file=env_plugmeter -p 3000:3000  -v ./out/:/out plugmeter
```

```
docker run --name plugmeter   --env-file=env_plugmeter -p 3000:3000  -v /var/plugmeter/out/:/out plugmeter
```


## Development

TODO

* CI 
* publish image on docker hub.