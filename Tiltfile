# Tiltfile for kasir-menu-service
docker_build('kasir-menu-service', '.')
k8s_yaml('k8s/deployment.yaml')
k8s_resource('kasir-menu-service', port_forwards=3002)