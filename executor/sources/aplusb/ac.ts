export {};

const inputText: string = await Deno.readTextFile("/dev/stdin");
const input = inputText.trim().split(/\s+/).map(Number);
console.log(input[0] + input[1]);
