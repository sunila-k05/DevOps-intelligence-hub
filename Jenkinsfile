pipeline {
    agent any

    environment {
        DOCKERHUB_USER = "sunilak05"
    }

    stages {

        stage('Checkout') {
            steps {
                git branch: 'main', url: 'https://github.com/sunila-k05/DevOps-intelligence-hub.git'
                script {
                    TAG = sh(script: "git rev-parse --short HEAD", returnStdout: true).trim()
                    echo "Using TAG: ${TAG}"
                }
            }
        }

        stage('Build Backend') {
            steps {
                script {
                    if (fileExists('backend/Dockerfile')) {
                        sh """
                        docker build -t ${DOCKERHUB_USER}/devops-intel-backend:${TAG} backend
                        """
                    } else {
                        echo "Backend Dockerfile not found"
                    }
                }
            }
        }

        stage('Build Frontend') {
            steps {
                script {
                    if (fileExists('frontend/Dockerfile')) {
                        sh """
                        docker build -t ${DOCKERHUB_USER}/devops-intel-frontend:${TAG} frontend
                        """
                    } else {
                        echo "Frontend Dockerfile not found"
                    }
                }
            }
        }

        stage('DockerHub Login') {
            steps {
                withCredentials([usernamePassword(
                    credentialsId: 'dockerhub-cred',
                    usernameVariable: 'DOCKER_USER',
                    passwordVariable: 'DOCKER_PASS'
                )]) {
                    sh """
                    echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin
                    """
                }
            }
        }

        stage('Push Images') {
            steps {
                sh "docker push ${DOCKERHUB_USER}/devops-intel-backend:${TAG}"
                sh "docker push ${DOCKERHUB_USER}/devops-intel-frontend:${TAG}"
            }
        }

        stage('Deploy to EC2') {
            steps {
                sshagent(credentials: ['ec2-ssh']) {
                    sh """
                    ssh -o StrictHostKeyChecking=no ubuntu@3.7.45.192'
                        docker pull ${DOCKERHUB_USER}/devops-intel-backend:${TAG}
                        docker pull ${DOCKERHUB_USER}/devops-intel-frontend:${TAG}

                        docker stop devops-backend || true
                        docker rm devops-backend || true
                        docker stop devops-frontend || true
                        docker rm devops-frontend || true

                        docker run -d --name devops-backend -p 8081:8081 ${DOCKERHUB_USER}/devops-intel-backend:${TAG}
                        docker run -d --name devops-frontend -p 80:80 ${DOCKERHUB_USER}/devops-intel-frontend:${TAG}
                    '
                    """
                }
            }
        }
    }

    post {
        success {
            echo "SUCCESS: Deployment finished with TAG ${TAG}"
        }
        failure {
            echo "BUILD FAILED — check logs"
        }
    }
}
