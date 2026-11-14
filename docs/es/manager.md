# Uso del gestor

Ejecuta `astroenvs` (o `ipynb`) en una terminal del servidor. Cada pregunta es
una lista que recorres con el teclado:

| Tecla | Acción |
|---|---|
| ↑ / ↓ | moverse |
| Enter | elegir |
| Esc o Ctrl+C | volver (en el menú principal: salir) |
| / | filtrar la lista escribiendo |
| Tab | completar el nombre de una carpeta o archivo, en los campos de texto |

Los servidores siguen corriendo aunque cierres el gestor o te desconectes.
Vuelve a abrir el gestor cuando quieras para verlos, leer sus logs o detenerlos.

## Iniciar Jupyter

**Jupyter · Start server** hace tres preguntas:

1. **Start in**: la carpeta donde se abre Jupyter. Tu carpeta personal aparece
   primero y es la opción por defecto; después vienen las carpetas de
   `DATA_DIRS` y **Custom path…** para cualquier otra.
2. **Access**: **Normal** llega a todos tus archivos; **Sandboxed** solo ve la
   carpeta del paso 1. Mira [Servidores aislados](#servidores-aislados).
3. **Token**: el token por defecto de `~/.astroenvs.env`, o **Type a different
   token…** para usar otro solo en este servidor.

Luego muestra la dirección que debes abrir:

```text
✔ Jupyter started in the background (PID 48211, port 8888)!
Open in your browser (VPN on):
  http://10.254.0.10:8888/lab?token=...
```

En un servidor normal, la carpeta que eliges solo es donde parte el explorador
de archivos de Jupyter y la más alta que muestra. La terminal y los notebooks de
Jupyter corren con tu cuenta y llegan a todos tus archivos, así que no compartas
la URL de un servidor normal.

Cada servidor usa el primer puerto libre desde el 8888. El servidor es
compartido, así que otra persona puede estar usando ya el 8888; el gestor lo
revisa y pasa al siguiente.

## Elegir una carpeta

La lista de carpetas marca con `(not found)` las que no existen en esta máquina
y no te deja elegirlas. Con **Custom path…** escribes una ruta:

- `~` es tu carpeta personal; las rutas relativas parten desde donde abriste el gestor
- **Tab** completa nombres de carpetas, como en la terminal
- Enter revisa que la carpeta exista antes de seguir

## Iniciar CARTA

**CARTA · Start server** pide la carpeta y el token, igual que Jupyter, y
muestra la URL. Los puertos de CARTA parten en el 3002.

La carpeta es solo donde se abre el explorador de archivos de CARTA. Desde ahí
puede subir hasta la carpeta definida en `CARTA_TOP` (`/` por defecto); ver
[Configuración](configuration.md).

## Instancias activas

**Jupyter · Active instances** y **CARTA · Active instances** listan los
servidores que están corriendo, con su carpeta y la URL completa, token
incluido. El menú principal también muestra cuántos hay, por ejemplo
`(2 running)`.

También aparecen los servidores de Jupyter que no inició este gestor (a mano, o
con el gestor en bash usado antes de la 1.0.0), con lo que muestra
`jupyter server list`.

## Logs

**Jupyter · Logs** y **CARTA · Logs** listan los archivos de log de esa
herramienta, del más reciente al más antiguo, y marcan los que corresponden a un
servidor que sigue corriendo. Elige uno para ver sus últimas 30 líneas. Si un
servidor no logra iniciar, el gestor muestra el final de su log en ese momento.

## DS9 y CASA

Las ventanas de DS9 y el logger y los gráficos de CASA son ventanas de
escritorio, así que necesitan reenvío X11: la ventana corre en el servidor y se
dibuja en tu computador.

En Mac, instala [XQuartz](https://www.xquartz.org), cierra sesión y vuelve a
entrar, y luego conéctate con:

```bash
ssh -Y <usuario>@10.254.0.10
```

En Linux, `ssh -Y` funciona tal cual. En Windows, usa un servidor X como VcXsrv
o MobaXterm.

**DS9 · Open window** pide una carpeta y luego un archivo FITS: lista los
archivos `.fits`, `.fit`, `.fts`, `.fits.gz` y `.fz` de esa carpeta, además de
**No file** y **Other file…**. DS9 se abre en su propia ventana y sigue
corriendo en segundo plano; su diálogo File parte en la carpeta que elegiste.

**CASA · Start in this terminal** ejecuta CASA en la terminal actual, dentro de
la carpeta que elegiste (CASA deja ahí sus archivos `casa-*.log`). Escribe
`exit` para volver al menú. Sin X11, CASA parte con `--nologger --nogui`: las
tareas y herramientas funcionan, solo faltan las ventanas.

## Detener

**Manage · Stop an instance** lista todo lo que está corriendo y detiene lo que
elijas:

```text
┃ Instance to stop
┃ > jupyter  PID 48211   port 8888  ~/Tareas/T5/data
┃   carta    PID 48302   port 3002  /data
┃   ds9      PID 48390   ~/data  m31.fits
┃   Cancel
```

**Manage · Stop all instances** detiene todo, o todos los servidores de un tipo.

Jupyter recibe una señal de cierre normal para que pueda cerrar sus kernels.
CARTA y DS9 se detienen junto con los procesos que hayan abierto.

## Servidores aislados

Elige **Sandboxed** cuando otra persona vaya a usar la URL, por ejemplo para
trabajar en `~/shared/T5` con un compañero. Dentro de un servidor aislado, los
notebooks y la terminal funcionan como siempre, pero solo existe esto:

| Dentro del sandbox | Acceso |
|---|---|
| La carpeta que elegiste | lectura y escritura |
| `~/.ipynb`, su Python y `~/.TinyTeX` | solo lectura |
| Carpetas del sistema (`/usr`, `/etc`, …) | solo lectura |
| Tu carpeta personal y `/tmp` | vacías, y se borran al detener el servidor |
| El resto de tus archivos, otras carpetas de datos | no existen |
| Tus otros procesos | invisibles |

Así, `cd ..` se detiene en la carpeta y `ls ~` no muestra nada tuyo. Los
paquetes instalados en `~/.ipynb` funcionan, pero nadie puede instalar en él ni
cambiarlo desde el sandbox.

Lo que no cambia:

- **Los archivos de la carpeta** los puede leer, cambiar y borrar quien tenga la
  URL. Revisa los scripts que aparezcan ahí antes de ejecutarlos fuera del
  sandbox.
- **La red**: el código dentro del sandbox igual puede llegar a internet y a
  otros servicios del servidor.
- **Los límites de tu cuenta**: cuenta para tu cuota de disco y CPU, y aparece
  como procesos tuyos.

No puedes aislar tu carpeta personal completa, ni `/`, porque contienen todo lo
demás. Elige una subcarpeta. El sandbox necesita Linux con user namespaces, que
el servidor del laboratorio tiene; en otros sistemas el gestor no lo ofrece. En
**Active instances** y **Stop an instance**, los servidores aislados aparecen
marcados `(sandboxed)`.

## Compartir y seguridad

Una URL de Jupyter con su token vale lo mismo que tu contraseña. En un servidor
normal, quien la abra tiene una terminal y notebooks que corren como tú: puede
leer, cambiar y borrar cualquier cosa de tu cuenta, sin importar en qué carpeta
iniciaste Jupyter. Una URL de CARTA da acceso a las imágenes que CARTA puede
abrir.

- **Comparte solo URLs de servidores aislados**, y solo con personas a las que
  dejarías cambiar esa carpeta.
- **Usa un token largo y aleatorio.** Los servidores se pueden alcanzar desde
  toda la red, así que una palabra como `Capybara` se puede adivinar. El gestor
  avisa si el token tiene menos de 16 caracteres. Para generar uno:
  `python3 -c "import secrets; print(secrets.token_urlsafe(24))"`
- **Detén un servidor compartido** cuando termines, o inicia uno nuevo con otro
  token, para cortar el acceso.

El gestor mantiene los tokens fuera del alcance de otros usuarios del servidor.
Le pasa el token a Jupyter por el entorno (cualquier usuario ve las líneas de
comando con `ps`) y deja `~/.astrolab`, que guarda los logs y las URLs, legible
solo para ti.

## Línea de comandos

```bash
astroenvs --version   # muestra la versión
astroenvs --help      # uso resumido
```
