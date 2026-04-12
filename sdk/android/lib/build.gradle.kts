// PulseSDK Android — Gradle build configuration
plugins {
    id("com.android.library")
    id("org.jetbrains.kotlin.android")
    `maven-publish`
}

android {
    namespace = "br.com.iit.pulse"
    compileSdk = 34

    defaultConfig {
        minSdk = 21
        targetSdk = 34
        version = "0.1.0"
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_1_8
        targetCompatibility = JavaVersion.VERSION_1_8
    }
    kotlinOptions {
        jvmTarget = "1.8"
    }
}

dependencies {
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3")
}

publishing {
    publications {
        create<MavenPublication>("release") {
            groupId = "br.com.institutoitinerante"
            artifactId = "pulse-sdk"
            version = "0.1.0"
        }
    }
}
