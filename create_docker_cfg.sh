cat <<EOF  > ~/.dockercfg
{ 
	"https://index.docker.io/v1/": { 
		"auth": "$AUTH", 
		"email": "$EMAIL" 
	} 
} 
EOF
