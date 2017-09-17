node {
	checkout scm

	stage('Test') {
		sh 'make test'
	}

	state('Build') {
		sh 'make build-provisioner publish-provisioner'
	}
}
