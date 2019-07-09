# PlanetLab Healthcheck script
# run as ./healthcheck [port] [response]
# defaults to ./healthcheck 9876 OK

# kill possible nc processes
pkill -f nc

# check for Internet access
wget -q --tries=10 --timeout=20 --spider https://images.ctfassets.net/2o3iq74rr1u2/6moBdVe2UwG0Fua6LTDQ6m/78cf10a8560cec2eb1009589ba6536f6/axel_profile_resized.jpg > /dev/null
if [[ $? -eq 0 ]]; then
        echo "Host appears to be online"
else
        echo "Host appears to be offline. Aborting."
        exit 1
fi

if [ -x "$(command -v python)" ]; then
        python -m SimpleHTTPServer 9876

elif ! [ -x "$(command -v ncat)" ]; then
        sudo yum install nmap -y --nogpgcheck
        # Listen for connections on external interface
        echo "Listening for connections on ${1:-9876}"
        RESPONSE="HTTP/1.1 200 OK\r\nConnection: keep-alive\r\n\r\n${2:-"OK"}\r\n"
        while { echo -en "$RESPONSE"; } | ncat -l 0.0.0.0 "${1:-9876}"; do
        echo "================================================"
        done
fi