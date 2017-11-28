# Templates

## Templates Env

In your `/etc/mayu/config.yaml` you can define template environment variables
that are dynamically usable in `templates/ignition.yaml` when
doing tne following.

```yaml
templates_env:
  my_var: '12345'
```

This sets the variable with key `my_var` to the corresponding
value within the template file. See https://golang.org/pkg/text/template/ for
more details about the usage of template variables in golang templates. Using
the injected configuration within your `templates/ignition.yaml`
looks like the following. Because of that you can configure ignition dynamically to your own needs.

```nohighlight
systemd:
  units:
    - name: extra_unit.service
      enable: true
      contents: |
        [Unit]
        Description=extra_unit
        [Service]
        Type=oneshot
        RemainAfterExit=yes
        ExecStart=/usr/bin/bash -c 'echo "{{index .TemplatesEnv "my_var"}}" >> /extra-file'
        [Install]
        WantedBy=multi-user.target

``

## Files
Ignition requires files to be specified via [data url format](https://tools.ietf.org/html/rfc2397). Which means no plaintext files in the ignition (like it was in cloudconfig).

Mayu is doing encoding of the files automatically, but they have to be seaprate files in the `./files` folder in the mayu directory.

Lets say that we want to have config file for `my-service` located on path `/etc/my-service/config.ini`. First you need to put file with the required content into `./files/my-service/config.ini`
```
[section1]
value1=test
value2=12345
value3=&^&HG
```

Then in ignition add this configuration:
```
storage:
  files:
    - filesystem: root
      path: /etc/my-service/config.conf
      mode: 0600
      user:
        id: 0
        group: 0
      contents:
        source:
          scheme: data
          opaque: "text/plain;charset=utf-8;base64,{{  index .Files "my-service/config.ini" }}"

```

Where `{{  index .Files "my-service/config.ini" }}` is definning which files should be put there. `my-service/config.ini` is the relative path to the file in the `./files` directory.
