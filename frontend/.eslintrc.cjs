module.exports = {
    extends: ['eslint:recommended', 'plugin:@typescript-eslint/recommended'],
    parser: '@typescript-eslint/parser',
    plugins: ['@typescript-eslint'],
    root: true,
    ignorePatterns: [
        "src/api/library_checker.ts",
        "src/api/library_checker.client.ts",
        "src/api/google/*"
    ]
};
