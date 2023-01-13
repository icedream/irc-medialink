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
	{{- reset }}

	»

	{{- if index . "AgeRestriction" }}
		{{ color 4 -}}
		{{ bold -}}
		[{{- index . "AgeRestriction" }}]
		{{- reset }}
	{{- end }}

	{{- if index . "IsLive" }}
		{{ bcolor 0 4 -}}
		{{ bold -}}
		[● LIVE]
		{{- reset }}
	{{- end }}

	{{- if index . "IsUpcomingLive" }}
		{{ bcolor 0 14 -}}
		{{ bold -}}
		[● LIVE]
		{{- reset }}
	{{- end }}

	{{- if index . "IsFinishedLive" }}
		{{ color 14 -}}
		[● FINISHED]
		{{- reset }}
	{{- end }}

	{{ if index . "IsProfile" }}
		{{- if index . "Title" }}
			{{ bold -}}
			{{- index . "Title" }}
			{{- if or (index . "Name") (index . "Description") -}}
			:
			{{- end -}}
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

		{{ with index . "Description" }}
			{{ excerpt 128 . }}
		{{ end }}
	{{ else }}
		{{ if index . "Title" }}
			{{ excerpt 184 (index . "Title") }}
			{{ if index . "IsUpcomingLive" }}
				{{ if index . "ScheduledStartTime" }}
					{{ with index . "ScheduledStartTime" }}
						(coming up {{ ago . }})
					{{ end }}
				{{ else }}
					(coming up)
				{{ end }}
			{{ end }}
			{{ if index . "IsLive" }}
				{{ if index . "ActualStartTime" }}
					{{ with index . "ActualStartTime" }}
						(started {{ ago . }})
					{{ end }}
				{{ else }}
					{{ with index . "Duration" }}
						({{ . }})
					{{ end }}
				{{ end }}
			{{ else }}
				{{ with index . "Duration" }}
					({{ . }})
				{{ end }}
			{{ end }}
		{{ else }}
			{{ with index . "Description" }}
				{{ excerpt 384 . }}
			{{ end }}
		{{ end }}
		
		{{ if index . "ImageType" }}
			{{ if index . "Title" }}
				·
			{{ end }}
			{{ .ImageType }} image,
			{{ if or (index . "ImageSize") (index . "Size") }}
				{{ with index . "ImageSize" }}
					{{ .X }}×{{ .Y }}
				{{ end }}
				{{ with index . "Size" }}
					({{ size . }})
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

	{{ if index . "Subscribers" }}
		·
		{{ with index . "Subscribers" }}
			👥{{ compactnum . }}
		{{ end }}
	{{ end }}

	{{ if or (index . "Likes") (or (index . "Favorites") (or (index . "Dislikes") (index . "Upvotes"))) }}
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
		{{ with index . "Upvotes" }}
			{{ color 7 -}}
			⬆️{{ compactnum . }}
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
		{{ with index . "Viewers"}}
			👥{{ compactnum . }}
		{{ end }}
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
