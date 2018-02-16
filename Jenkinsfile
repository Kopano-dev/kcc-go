#!/usr/bin/env groovy

pipeline {
	agent {
		docker {
			image 'golang:1.9'
			args '-u 0'
		}
	}
	environment {
		GLIDE_VERSION = 'v0.13.1'
		GLIDE_HOME = '/tmp/.glide'
		GOBIN = '/usr/local/bin'
		GOPATH = '/workspace'
		PACKAGE = 'stash.kopano.io/kgol/kcc-go'
	}
	stages {
		stage('Bootstrap') {
			steps {
				sh 'go env'
				sh 'export'
				echo 'Bootstrapping..'
				sh 'mkdir -p \$GOPATH/src/\$PACKAGE && rmdir \$GOPATH/src/\$PACKAGE && ln -sv \$WORKSPACE \$GOPATH/src/\$PACKAGE'
				sh 'curl -sSL https://github.com/Masterminds/glide/releases/download/$GLIDE_VERSION/glide-$GLIDE_VERSION-linux-amd64.tar.gz | tar -vxz -C /usr/local/bin --strip=1'
				sh 'go get -v github.com/golang/lint/golint'
			}
		}
		stage('Lint') {
			steps {
				echo 'Linting..'
				sh 'cd \$GOPATH/src/\$PACKAGE && golint \$(glide nv) | tee golint.txt || true'
				sh 'cd \$GOPATH/src/\$PACKAGE && go vet \$(glide nv) | tee govet.txt || true'
			}
		}
	}
	post {
		always {
			warnings parserConfigurations: [[parserName: 'Go Lint', pattern: 'golint.txt'], [parserName: 'Go Vet', pattern: 'govet.txt']], unstableTotalAll: '0', messagesPattern: 'don\'t use ALL_CAPS in Go names; use CamelCase'
			cleanWs()
		}
	}
}
