package source.defra;

public class DefraWrapper {
    static {
        System.loadLibrary("defradb");
        System.loadLibrary("nativewrapper");
    }

    // Native method declarations (private)
    private static native int initNodeNative(String dbPath);
    private static native int startNodeNative();
    private static native int stopNodeNative();
    private static native int addSchemaNative(String schema);
    private static native int addDocumentNative(String collection, String json);
    private static native int deleteDocumentNative(String collection, String docID);
    private static native String executeQueryNative(String query);

    // Public wrappers
    public int initNode(String dbPath) {
        return initNodeNative(dbPath);
    }

    public int startNode() {
        return startNodeNative();
    }

    public int stopNode() {
        return stopNodeNative();
    }

    public int addSchema(String schema) {
        return addSchemaNative(schema);
    }

    public int addDocument(String collection, String json) {
        return addDocumentNative(collection, json);
    }

    public int deleteDocument(String collection, String docID) {
        return deleteDocumentNative(collection, docID);
    }

    public static String executeQuery(String query) {
        return executeQueryNative(query);
    }
}