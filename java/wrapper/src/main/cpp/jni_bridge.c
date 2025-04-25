#include <jni.h>
#include <string.h>
#include <stdlib.h>
#include "libdefradb.h"

JNIEXPORT jint JNICALL 
Java_source_defra_DefraWrapper_initNodeNative(JNIEnv *env, jobject thiz, jstring jpath) {
    const char *cPath = (*env)->GetStringUTFChars(env, jpath, 0);

    // Call the Go function
    int result = initNode((char *)cPath);

    (*env)->ReleaseStringUTFChars(env, jpath, cPath);
    return result;
}

JNIEXPORT jint JNICALL
Java_source_defra_DefraWrapper_startNodeNative(JNIEnv *env, jclass clazz) {
    return startNode();
}

JNIEXPORT jint JNICALL
Java_source_defra_DefraWrapper_stopNodeNative(JNIEnv *env, jclass clazz) {
    return stopNode();
}

JNIEXPORT jint JNICALL
Java_source_defra_DefraWrapper_addSchemaNative(JNIEnv *env, jclass clazz, jstring schema) {
    const char *cSchema = (*env)->GetStringUTFChars(env, schema, 0);
    jint result = addSchema((char *)cSchema);
    (*env)->ReleaseStringUTFChars(env, schema, cSchema);
    return result;
}

JNIEXPORT jint JNICALL
Java_source_defra_DefraWrapper_addDocumentNative(JNIEnv *env, jclass clazz, jstring collection, jstring json) {
    const char *cCollection = (*env)->GetStringUTFChars(env, collection, 0);
    const char *cJSON = (*env)->GetStringUTFChars(env, json, 0);
    jint result = addDocument((char *)cCollection, (char *)cJSON);
    (*env)->ReleaseStringUTFChars(env, collection, cCollection);
    (*env)->ReleaseStringUTFChars(env, json, cJSON);
    return result;
}

JNIEXPORT jint JNICALL
Java_source_defra_DefraWrapper_deleteDocumentNative(JNIEnv *env, jclass clazz, jstring collection, jstring docID) {
    const char *cCollection = (*env)->GetStringUTFChars(env, collection, 0);
    const char *cDocID = (*env)->GetStringUTFChars(env, docID, 0);
    jint result = deleteDocument((char *)cCollection, (char *)cDocID);
    (*env)->ReleaseStringUTFChars(env, collection, cCollection);
    (*env)->ReleaseStringUTFChars(env, docID, cDocID);
    return result;
}

JNIEXPORT jstring JNICALL
Java_source_defra_DefraWrapper_executeQueryNative(JNIEnv *env, jclass clazz, jstring query) {
    const char *cQuery = (*env)->GetStringUTFChars(env, query, 0);
    char *result = executeQuery((char *)cQuery);
    (*env)->ReleaseStringUTFChars(env, query, cQuery);
    jstring jResult = (*env)->NewStringUTF(env, result);
	free(result);
    return jResult;
}