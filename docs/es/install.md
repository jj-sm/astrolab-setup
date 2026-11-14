# Instalación

## Requisitos

- Una cuenta en un servidor Linux con acceso a internet (el servidor del
  laboratorio, o cualquier otra máquina a la que entres por SSH)
- `git`, `curl` y `wget`, que el servidor ya tiene
- Cerca de 1 GB libre en tu carpeta personal para Python, el entorno virtual y
  TinyTeX

No necesitas `sudo`: todo queda en tu carpeta personal.

## Instalación rápida

En el servidor, ejecuta:

```bash
curl -fsSL https://raw.githubusercontent.com/jj-sm/astrolab-setup/main/install.sh | bash
```

Esto hace por ti los pasos 1 a 3: descarga el repositorio en una carpeta
temporal, te pide elegir un token, ejecuta la instalación y borra la carpeta, así
que en tu carpeta personal solo queda lo que instala la instalación. No necesita
`git`. Presiona Enter
cuando pida el token para usar uno aleatorio; se muestra en pantalla y queda
guardado en `~/.astroenvs.env`. Si ya definiste un token en una instalación
anterior, lo conserva y no pregunta, junto con el resto de tu configuración en
`~/.astroenvs.env`.

Cuando termine, cierra y vuelve a abrir la terminal y sigue con el paso 4.
Ejecutar el mismo comando más adelante actualiza todo.

Si antes instalaste a mano, después puedes borrar `~/astrolab-setup`: nada lo usa
una vez que la instalación terminó.

El resto de esta página hace lo mismo a mano.

## 1. Descarga el repositorio

Clónalo en tu carpeta personal. La instalación lee `.env` y `folder.bash` de la
carpeta del repositorio, desde donde sea que la ejecutes.

```bash
cd ~
git clone https://github.com/jj-sm/astrolab-setup.git
```

## 2. Elige un token

Jupyter y CARTA piden un token antes de dejar entrar a alguien. Define uno por
defecto en un archivo `.env`:

```bash
cp astrolab-setup/.env.example astrolab-setup/.env
nano astrolab-setup/.env        # define TOKEN=<algo aleatorio>
```

Usa solo letras, dígitos y `. _ ~ -`, porque el token va dentro de URLs. Para
generar uno aleatorio: `python3 -c "import secrets; print(secrets.token_urlsafe(24))"`.

La instalación copia este archivo a `~/.astroenvs.env` con permisos `600`, así
que solo tú puedes leerlo. git ignora `.env`; nunca lo subas al repositorio.

Puedes saltarte este paso: en ese caso el gestor te pedirá escribir un token
cada vez que inicies un servidor.

## 3. Ejecuta la instalación

```bash
bash astrolab-setup/ipynb_setup.bash
```

Hace estos pasos y se detiene en el primer error:

1. Instala `uv`
2. Instala Python 3.11 y 3.13 con `uv`
3. Muestra las versiones instaladas
4. Crea el entorno virtual `~/.ipynb` con Python 3.13, o lo reutiliza si ya
   existe, conservando los paquetes que instalaste en él
5. Instala `ipykernel` e `ipynb-jjsm-tools[full]` en él y registra el kernel
   *Python (ipynb 3.13)*
6. Lista los paquetes instalados
7. Instala TinyTeX, si no está ya
8. Instala los paquetes de LaTeX que necesita Matplotlib, si encuentra `tlmgr`
9. Descarga la última versión de `astroenvs` en `~/.local/bin/astroenvs`, copia
   `.env` a `~/.astroenvs.env` y `folder.bash` a `~/.folder.sh`, y agrega los
   alias a `~/.bashrc`, `~/.bash_profile` y `~/.profile`

Si la descarga del paso 9 falla, por ejemplo porque no se puede llegar a
GitHub, la instalación igual termina el resto del paso y luego se detiene con un
error que muestra el comando `curl` para descargar el gestor a mano. Los alias ya
apuntan a `~/.local/bin/astroenvs`, así que el gestor funciona apenas lo
descargues.

Cuando termine, carga los alias nuevos:

```bash
source ~/.bashrc
```

## 4. Comprueba que funciona

```bash
astroenvs --version     # astroenvs v1.0.1
astroenvs               # abre el menú
```

## Actualizar el gestor

No hace falta repetir toda la instalación. Descarga la última versión sobre el
binario anterior:

```bash
curl -fsSL https://github.com/jj-sm/astrolab-setup/releases/latest/download/astroenvs-linux-amd64 \
  -o ~/.local/bin/astroenvs && chmod +x ~/.local/bin/astroenvs
```

En un servidor ARM, cambia `linux-amd64` por `linux-arm64`.

Volver a ejecutar el script de instalación también es seguro: reinstala el
gestor y reescribe los alias sin duplicarlos.

Si instalaste antes de la versión 1.0.0, volver a ejecutar la instalación también
cambia los alias `astroenvs` e `ipynb` del antiguo `bash ~/.astroenvs.sh` al
binario nuevo. Después puedes borrar `~/.astroenvs.sh`.

## Carpetas de proyecto (`tfold`)

`tfold` crea la estructura habitual de un proyecto nuevo en la carpeta actual:

```bash
mkdir ~/mi-proyecto && cd ~/mi-proyecto
tfold          # crea out/ src/ data/ notebooks/, cada una con un .gitkeep
```
