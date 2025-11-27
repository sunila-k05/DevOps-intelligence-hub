pipeline {
    agent any

    environment {
        DOCKERHUB_USER = 'sunilak05'
        DOCKERHUB_CRED = credentials('dockerhub-cred')   // your Jenkins credential ID
        GITHUB_CRED = credentials('github')              // your GitHub token
    }

    stages {

        /* ---------------------- CHECKOUT CODE ---------------------- */
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: '*/main']],
                    userRemoteConfigs: [[
                        url: 'https://github.com/sunila-k05/DevOps-intelligence-hub.git',
                        credentialsId: 'github'
                    ]]
                ])

                script {
                    TAG = sh(script: "git rev-parse --short HEAD", returnStdout: true).trim()
                    echo "Using TAG: ${TAG}"
                }
            }
        }

        /* ---------------------- PRE-PULL DOCKER IMAGES ---------------------- */
        stage('Preload Base Images') {
            steps {
                sh '''
                    echo "Pre-pulling base images (cache boost)..."
                    docker pull golang:1.22-alpine || true
                    docker pull alpine:3.20 || true
                    docker pull node:20-alpine || true
                    docker pull nginx:alpine || true
                '''
            }
        }

        /* ---------------------- BUILD BACKEND ---------------------- */
        stage('Build Backend') {
            steps {
                script {
                    if (fileExists('backend/Dockerfile')) {
                        echo "Building backend image..."
                        sh """
                            docker build -t ${DOCKERHUB_USER}/devops-intel-backend:${TAG} backend
                        """
                    } else {
                        echo "Backend Dockerfile missing → skipping backend build"
                    }
                }
            }
        }

        /* ---------------------- BUILD FRONTEND ---------------------- */
        stage('Build Frontend') {
            steps {
                script {
                    if (fileExists('frontend/Dockerfile')) {
                        echo "Building frontend image..."
                        sh """
                            docker build -t ${DOCKERHUB_USER}/devops-intel-frontend:${TAG} frontend
                        """
                    } else {
                        echo "Frontend Dockerfile missing → skipping frontend build"
                    }
                }
            }
        }

        /* ---------------------- DOCKER LOGIN ---------------------- */
        stage('Docker Login') {
            steps {
                sh """
                    echo ${DOCKERHUB_CRED_PSW} | docker login -u ${DOCKERHUB_CRED_USR} --password-stdin
                """
            }
        }

        /* ---------------------- PUSH IMAGES ---------------------- */
        stage('Push Images') {
            steps {
                script {
                    if (sh(script: "docker image inspect ${DOCKERHUB_USER}/devops-intel-backend:${TAG}", returnStatus: true) == 0) {
                        sh "docker push ${DOCKERHUB_USER}/devops-intel-backend:${TAG}"
                    } else {
                        echo "Backend image not built → skipping push"
                    }

                    if (sh(script: "docker image inspect ${DOCKERHUB_USER}/devops-intel-frontend:${TAG}", returnStatus: true) == 0) {
                        sh "docker push ${DOCKERHUB_USER}/devops-intel-frontend:${TAG}"
                    } else {
                        echo "Frontend image not built → skipping push"
                    }
                }
            }
        }

        /* ---------------------- OPTIONAL: DEPLOY TO EC2 ---------------------- */
        stage('Deploy to EC2 (Optional)') {
            when { expression { return false } }   // enable later
            steps {
                sh '''
                    echo "Deployment disabled. Enable this stage when ready."
                    # Example:
                    # docker pull sunilak05/devops-intel-backend:${TAG}
                    # docker pull sunilak05/devops-intel-frontend:${TAG}
                    # docker-compose down && docker-compose up -d
                '''
            }
        }
    }

    post {
        success {
            echo "Build Successful → Images pushed with tag: ${TAG}"
        }
        failure {
            echo "Build Failed ❌ — Check logs."
        }
    }
}
