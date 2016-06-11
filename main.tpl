{{ define "verified" }}
	{{ bold -}}
	{{ bcolor 0 12 -}}
	✔
	{{- reset }}
{{ end }}

{{ define "error" }}
	{{ bold -}}
	{{ color 4 -}}
	ERROR:
	{{- reset }}
	
	{{ . }}
{{ end }}

{{ define "link-info" }}
	{{ bold -}}
	{{- if index . "Header" -}}
		{{- index . "Header" -}}
	{{- else -}}
		Link info
	{{- end -}}
	{{- bold }}

	»
	
	{{ if index . "IsProfile" }}
		{{- if index . "Title" }}
			{{ bold -}}
			{{- index . "Title" }}:
			{{- bold }}
		{{- end }}

		{{ if index . "Name" }}
			{{ excerpt 184 (index . "Name") }}
			{{ if index . "Verified" }}
				{{ template "verified" }}
			{{ end }}
			
			{{ if or (index . "CountryCode") (index . "City") }}
				from
				{{ if and (index . "CountryCode") (index . "City") }}
					{{ index . "City" }},
					{{ index . "CountryCode" }}
				{{ else }}
					{{ with index . "City" }}
						{{ . }}
					{{ end }}
					{{ with index . "CountryCode" }}
						{{ . }}
					{{ end }}
				{{ end }}
			{{ end }}
		{{ end }}
	{{ else }}
		{{ if index . "Title" }}
			{{ excerpt 184 (index . "Title") }}
			{{ with index . "Duration" }}
				({{ . }})
			{{ end }}
		{{ else }}
			{{ if index . "Description" }}
				{{ excerpt 384 (index . "Description") }}
			{{ else }}
				{{ with index . "ImageType" }}
					{{ . }} image,
				{{ end }}
				{{ if (index . "ImageSize") (index . "Size") }}
					{{ with index . "ImageSize" }}
						{{ .X }}x{{ .Y }}
					{{ end }}
					{{ with index . "Size" }}
						({{ size . }})
					{{ end }}
				{{ end }}
			{{ end }}
		{{ end }}
	{{ end }}

	{{ if or (index . "Author") }}
		{{ if index . "Author" }}
			{{ with index . "Author" }}
				by {{ excerpt 184 . }}
			{{ end }}
			{{ if index . "AuthorIsVerified" }}
				{{ template "verified" }}
			{{ end }}
		{{ end }}
	{{ end }}
	
	{{ if index . "Followers" }}
		·
		{{ with index . "Followers" }}
			👥{{ compactnum . }}
		{{ end }}
	{{ end }}

	{{ if or (index . "Likes") (or (index . "Favorites") (index . "Dislikes")) }}
		·
		{{ with index . "Likes" }}
			{{ color 3 -}}
			👍{{ compactnum . }}
			{{- reset }}
		{{ end }}
		{{ with index . "Dislikes" }}
			{{ color 4 -}}
			👎{{ compactnum . }}
			{{- reset }}
		{{ end }}
		{{ with index . "Favorites" }}
			{{ color 7 -}}
			❤{{ compactnum . }}
			{{- reset }}
		{{ end }}
		{{ with index . "Reposts" }}
			{{ color 12 -}}
			🔁{{ compactnum . }}
			{{- reset }}
		{{ end }}
	{{ end }}
	
	{{ if or (index . "Views") (or (index . "Plays") (or (index . "Downloads") (or (index . "Uploads") (index . "Comments")))) }}
		· 
		{{ with index . "Views" }}
			👁{{ compactnum . }}
		{{ end }}
		{{ with index . "Plays" }}
			▶{{ compactnum . }}
		{{ end }}
		{{ with index . "Downloads" }}
			⬇{{ compactnum . }}
		{{ end }}
		{{ with index . "Uploads" }}
			⬆️{{ compactnum . }}
		{{ end }}
		{{ with index . "Comments" }}
			💬{{ compactnum . }}
		{{ end }}
	{{ end }}
{{ end }}
