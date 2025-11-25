pipeline {
    agent any

    environment {
        REGISTRY = "your-dockerhub-username"   // <- change this
        IMAGE = "devops-intel-hub"             // <- change if needed
    }

    stages {
        stage('Checkout') {
            steps {
                git branch: 'main',
                    url: 'https://github.com/sunila-k05/DevOps-intelligence-hub.git',
                    credentialsId: 'github'
            }
        }

        stage('Build / Test') {
            steps {
                // run build commands; adapt for your project
                // e.g.: for maven: sh 'mvn -B -DskipTests=false test package'
                echo "Add your build/test commands here"
            }
        }

        stage('Build Docker Image') {
            steps {
                script {
                    // Builds image and stores reference in variable
                    dockerImage = docker.build("${REGISTRY}/${IMAGE}:${env.BUILD_NUMBER}")
                }
            }
        }

        stage('Login to Docker Hub') {
            steps {
                withCredentials([usernamePassword(credentialsId: 'dockerhub',
                                                  usernameVariable: 'DOCKER_USER',
                                                  passwordVariable: 'DOCKER_PASS')]) {
                    sh 'echo $DOCKER_PASS | docker login -u $DOCKER_USER --password-stdin'
                }
            }
        }

        stage('Push Docker Image') {
            steps {
                script {
                    dockerImage.push()
                    // Optionally push 'latest' tag too:
                    dockerImage.push('latest')
                }
            }
        }

        stage('Deploy (placeholder)') {
            steps {
                echo "Deploy stage — customize (ssh to server, kubectl apply, etc.)"
            }
        }
    }

    post {
        always {
            // cleanup
            sh 'docker image prune -f || true'
        }
    }
}
