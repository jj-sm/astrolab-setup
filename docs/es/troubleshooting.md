# Solución de problemas

## `astroenvs: command not found`

Los alias se cargan en las terminales nuevas. Ejecuta `source ~/.bashrc`, o
cierra sesión y vuelve a entrar. Si sigue fallando, revisa que el binario esté:

```bash
ls -l ~/.local/bin/astroenvs
~/.local/bin/astroenvs --version
```

Si no está, descárgalo de nuevo (ver [Actualizar el gestor](install.md#actualizar-el-gestor)).

## La instalación dice que falló la descarga de astroenvs

La instalación descarga el gestor desde las releases de GitHub. Si eso falla, el
resto de la instalación igual se completa, y las últimas líneas muestran el
comando para descargarlo a mano. Revisa que el servidor pueda llegar a GitHub
(`curl -I https://github.com`) y luego ejecuta ese comando, o el de
[Actualizar el gestor](install.md#actualizar-el-gestor).

## "~/.ipynb exists but isn't a virtual environment"

La instalación encontró una carpeta `~/.ipynb` que no es un entorno de Python, y
se detuvo en vez de borrarla. Revisa qué tiene, muévela a otro lugar (o bórrala
si no la necesitas) y vuelve a ejecutar la instalación.

## "Could not find jupyter in ~/.ipynb/bin"

El gestor ejecuta Jupyter desde el entorno virtual `~/.ipynb`. O el entorno
está en otro lugar, o Jupyter no está instalado en él.

- Si el entorno está en otro lugar, indícaselo al gestor con
  `IPYNB_VENV=/ruta/al/venv` en `~/.astroenvs.env`.
- Si falta Jupyter, instálalo:

  ```bash
  source ~/.ipynb/bin/activate
  uv pip install jupyterlab ipykernel
  ```

## El navegador no abre la URL

- Activa la VPN de la universidad: el servidor no es accesible desde fuera.
- Revisa en **Active instances** que el servidor siga corriendo.
- Si la dirección de la URL no es la del servidor, define `SERVER_IP` en
  `~/.astroenvs.env` (ver [Configuración](configuration.md)).

## "Failed to start" al iniciar un servidor

El gestor muestra las últimas líneas del log. El log completo está en
`~/.astrolab/logs/` y también en **Logs** dentro del menú. Causas comunes:

- **Puerto ocupado**: alguien tomó el puerto entre la revisión y el inicio.
  Vuelve a iniciarlo; el gestor elige el siguiente puerto libre.
- **Falta un paquete o extensión**: instálalo en `~/.ipynb`, como arriba.

## "carta executable not found" (o ds9, casa)

La herramienta no está en tu `PATH`. Busca dónde está instalada y defínela en
`~/.astroenvs.env`:

```bash
CARTA_BIN=/ruta/a/carta
DS9_BIN=/ruta/a/ds9
CASA_BIN=/ruta/a/casa
```

## "No X display" en DS9, o CASA sin ventanas

Tu sesión SSH no tiene reenvío X11. Reconéctate con `ssh -Y`, con XQuartz
instalado si usas Mac (ver [DS9 y CASA](manager.md#ds9-y-casa)). Para
comprobarlo, ejecuta `echo $DISPLAY`: debería mostrar algo como `localhost:10.0`.

## La carpeta que quiero no está en la lista

Elige **Custom path…** y escríbela (Tab completa). Para que aparezca siempre,
agrégala a `DATA_DIRS` en `~/.astroenvs.env`.

## Un servidor sigue corriendo después de detenerlo

Revisa **Active instances** otra vez: Jupyter puede tardar unos segundos en
cerrar sus kernels. Si el proceso sigue ahí, búscalo y detenlo a mano:

```bash
pgrep -u $USER -fl 'jupyter-lab|carta|ds9'
kill <PID>
```

## No encuentra Python o uv después de la instalación

`uv` se instala en `~/.local/bin` (las versiones antiguas usaban
`~/.cargo/bin`). Asegúrate de que esa carpeta esté en tu `PATH` y revisa:

```bash
uv --version
uv python find 3.13
```
