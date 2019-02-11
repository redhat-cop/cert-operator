library identifier: "pipeline-library@master",
retriever: modernSCM(
  [
    $class: "GitSCMSource",
    remote: "https://github.com/redhat-cop/pipeline-library.git"
  ]
)

openshift.withCluster() {
  env.NAMESPACE = openshift.project()
  env.APP_NAME = "cert-operator"
  ARTIFACT_DIRECTORY = "build/bin"
}

pipeline {

  agent {
    label 'jenkins-slave-golang'
  }

   environment {
    GOPATH=${WORKSPACE}
   }

  stages {

    stage('Setup Jenkins Environment') {
      steps {
        sh """
          mkdir -p ${WORKSPACE}/cert-operator/src/github.com/redhat-cop
        """
      }
    }

  	stage('Git Checkout') {
      steps {
        git checkout scm
      }
    }

    stage('Build Cert Operator') {
      steps {
       	sh """
          echo $GOPATH
          pwd
          ls -al
          ./build.sh
        """
      }
  	}

  	stage('Build Image') {
      steps {
       	binaryBuild(projectName: "${NAMESPACE}", buildConfigName: "${APP_NAME}", artifactsDirectoryName: "${ARTIFACT_DIRECTORY}")
      }
    }
  }
}