module.exports = {
    extends: ['eslint:recommended', 'plugin:@typescript-eslint/recommended'],
    parser: '@typescript-eslint/parser',
    plugins: ['@typescript-eslint'],
    root: true,
    ignorePatterns: [
        "build",
        "src/api/Library_checkerServiceClientPb.ts",
        "**/*_pb.js",
        "**/*_pb.d.ts"
    ],
    rules: {
//        "@typescript-eslint/no-empty-interface" : "off",
    }
};
