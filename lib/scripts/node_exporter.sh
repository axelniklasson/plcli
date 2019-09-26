install_node_exporter () {
    cd
    wget https://github.com/prometheus/node_exporter/releases/download/v0.18.1/node_exporter-0.18.1.linux-amd64.tar.gz
    tar xvfz node_exporter-0.18.1.linux-amd64.tar.gz
    sudo mv node_exporter-0.18.1.linux-amd64/node_exporter /usr/bin/
    rm -rf node_exporter-0.18.1.linux-amd64*
}

if ! [ -x "$(command -v node_exporter)" ]; then
	install_node_exporter
else
	echo "node_exporter already installed, skipping"
fi

node_exporter --web.listen-address=$(curl ifconfig.me):2100