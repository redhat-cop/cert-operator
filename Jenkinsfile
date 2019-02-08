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
    label 'maven'
  }

  stages {
  	stage('Git Checkout') {
      steps {
        git url: "${SOURCE_REPOSITORY_URL}"
      }
    }

    stage('Build Cert Operator') {
      steps {
       	sh 'pwd && ./build.sh'
      }
  	}

  	stage('Build Image') {
      steps {
       	binaryBuild(projectName: "${NAMESPACE}", buildConfigName: "${APP_NAME}", artifactsDirectoryName: "${ARTIFACT_DIRECTORY}")
      }
    }
  }
}