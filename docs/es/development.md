# Desarrollo

## Estructura del repositorio

| Ruta | Contenido |
|---|---|
| `cmd/astroenvs/` | el gestor, en Go |
| `ipynb_setup.bash` | el script de instalación |
| `folder.bash` | crea las carpetas de un proyecto (se instala como `tfold`) |
| `docs/en/`, `docs/es/` | esta documentación, en inglés y español |
| `.github/workflows/build.yml` | compila los binarios y publica las versiones |
| `.github/workflows/wikijs-docs.yml` | publica `docs/` en docs.jjsm.science |

## El código del gestor

`cmd/astroenvs/` es un solo paquete de Go:

| Archivo | Contenido |
|---|---|
| `main.go` | flags y el menú principal |
| `config.go` | lectura de `~/.astroenvs.env` y valores por defecto |
| `ui.go` | las preguntas: selector de carpetas, token, autocompletado de rutas, estilos |
| `actions.go` | una función por opción del menú |
| `process.go` | inicio de procesos separados, registro de instancias, puertos, logs |
| `sandbox_linux.go` | servidores aislados (`sandbox_other.go`: el reemplazo para otros sistemas) |

Las preguntas usan [huh](https://github.com/charmbracelet/huh). Los servidores
parten en su propia sesión (`setsid`), así que sobreviven al cierre del gestor y
de la sesión SSH. Cada instancia se guarda como JSON en `~/.astrolab/run/`; los
archivos de texto que escribía el gestor en bash usado antes de la 1.0.0 se
siguen leyendo, así que los servidores que dejó corriendo aparecen.

Un servidor aislado funciona como [bubblewrap](https://github.com/containers/bubblewrap).
El gestor se vuelve a iniciar a sí mismo como `astroenvs __sandbox` en nuevos
namespaces de usuario, montajes y procesos. Ese proceso arma una raíz que solo
contiene las carpetas del sistema y el venv (montajes bind de solo lectura) y la
carpeta elegida, cambia a ella con `pivot_root`, descarta todas sus capabilities
e inicia Jupyter, quedándose como PID 1. Descartar las capabilities en todos los
hilos requiere compilar sin cgo (`CGO_ENABLED=0`); con cgo, el sandbox se niega
a partir en vez de correr sin aislamiento.

## Compilar

Necesitas Go 1.24 o más reciente.

```bash
go build -o astroenvs ./cmd/astroenvs
./astroenvs --version              # astroenvs dev
```

Para el servidor, compila un binario estático de Linux:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags "-s -w -X main.version=v1.0.0" -o astroenvs-linux-amd64 ./cmd/astroenvs
```

Ejecuta `go vet ./...` y `gofmt -l cmd/` antes de hacer commit.

## Pruebas

`go test ./...` prueba el sandbox en Linux. `TestSandbox` ejecuta una shell
dentro y revisa qué puede ver y cambiar. `TestSandboxJupyter` inicia un
JupyterLab aislado real e intenta salir de la carpeta desde una terminal de
Jupyter; necesita un venv con JupyterLab en `ASTROENVS_TEST_VENV`. Ambas se
saltan donde no hay user namespaces, salvo con `ASTROENVS_SANDBOX_TEST=1`. CI
las ejecuta en cada compilación y arma el venv con uv, igual que la instalación.

```bash
CGO_ENABLED=0 ASTROENVS_TEST_VENV=~/.ipynb go test -v ./...
```

## Versiones

`build.yml` compila binarios para `linux/amd64`, `linux/arm64`, `darwin/amd64` y
`darwin/arm64` en cada push y pull request, y los adjunta a la ejecución del
workflow. Al subir un tag que empiece con `v`, además los publica, junto con un
`checksums.txt`, como una GitHub Release:

```bash
git tag v1.1.0
git push origin v1.1.0
```

El script de instalación y el comando de actualización descargan desde
`releases/latest`, así que una versión nueva le llega a todos los que actualicen.

## Esta documentación

La documentación es Markdown en `docs/en/` y `docs/es/`. Los archivos con el
mismo nombre son traducciones entre sí y quedan enlazados en la wiki, así que
mantén ambos idiomas al día.

- `README.md` es la página de resumen en `/astrophysics/astrolab-setup`
- cada otro archivo `nombre.md` queda en `/astrophysics/astrolab-setup/nombre`,
  y en `/es/astrophysics/astrolab-setup/nombre` para español
- el primer `# Título` es el título de la página
- enlaza entre páginas con enlaces relativos (`[Configuración](configuration.md)`);
  se reescriben al idioma correcto
- `_sidebar.json` en cada carpeta es la barra lateral de ese idioma

`wikijs-docs.yml` la publica en cada push a `main` que cambie `docs/`, usando
`.github/scripts/wikijs_sync.py` y la entrada del proyecto en
`.github/wikijs-projects.json`. También se puede ejecutar a mano desde la
pestaña Actions.

Para previsualizar en local sin publicar:

```bash
python3 .github/scripts/wikijs_sync.py --src docs --locales en,es \
  --prefix astrophysics/astrolab-setup --dry-run
```

Esto escribe las páginas, tal como se publicarían, en `wikisync-preview/`.

El workflow necesita estos secrets del repositorio: `WIKI_USERNAME`,
`WIKI_PASSWORD`, `WIKI_SITE_ID`, `CF_API_TOKEN` y `CF_ZONE_ID`. Además, el grupo
del usuario de la wiki necesita una regla de páginas que le permita escribir
bajo `astrophysics/astrolab-setup`, en ambos idiomas.
