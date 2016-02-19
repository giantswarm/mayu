# Templates Env

In your `/etc/mayu/config.yaml` you can define template environment variables
that are dynamically usable in `templates/last_stage_cloudconfig.yaml` when
doing tne following.

```yaml
templates_env:
  ssh_authorized_keys:
    - "pub key one"
    - "pub key two"
```

This sets the variable with key `ssh_authorized_keys` to the corresponding
value within the template file. See https://golang.org/pkg/text/template/ for
more details about the usage of template variables in golang templates. Using
the injected configuration within your `templates/last_stage_cloudconfig.yaml`
looks like the following. Because the last stage template is effectively a
cloud config, that way you can configure it dynamically to your own needs. As
here seen configuring ssh keys.

```nohighlight
ssh_authorized_keys:
{{ range $index, $pubkey := (index .TemplatesEnv "ssh_authorized_keys")}}- {{ $pubkey }}
{{end}}
```
