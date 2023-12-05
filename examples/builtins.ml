puts("testing builtins\n");

let arr = ["a", 123, [1, 2]];
let str = "length7";

puts("length of string " + str + " is " + string(len(str)) + "\n");
puts("length of array " + string(arr) + " is " + string(len(arr)) + "\n");

let arr = [1, 2, 3, 4];

puts("push 5 to " + string(arr), push(arr, 5));

puts("pop 4 from " + string(arr));

pop(arr);

puts(string(arr));

let hash = {"a": 1, "b": 2, "c": 3};

puts("keys of " + string(hash) + " are", string(keys(hash)));

puts("delete a from " + string(hash));

delete(hash, "a");

puts(string(hash));

let str = "mystring";
let arr = [1, 2, 3, 4];
let hash = {"a": 1, "b": 2, "c": 3};

puts("assign c to index 3 in " + str);
assign(str, 3, "c");
puts(str);

puts("assign 5 to index 2 in " + string(arr));
assign(arr, 2, 5);
puts(arr);

puts("assign 5 to c in " + string(hash));
assign(hash, "c", 5);
puts(hash);

puts("type of " + str + " is " + type(str));
puts("type of " + string(arr) + " is " + type(arr));
puts("type of " + string(hash) + " is " + type(hash));

puts("writing 'example' to 'newfile.txt'");

let file = open("newfile.txt", "w");

write(file, "example");

close(file);

let cmd = command("ls");
puts(cmd["stdout"]);

let file = open("newfile.txt", "r");
puts("reading from 'newfile.txt'");
puts(read(file));

puts("removing 'newfile.txt'");
remove(file);

let cmd = command("ls");
puts(cmd["stdout"]);
