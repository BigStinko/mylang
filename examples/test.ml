let input = open("/dev/stdin");
let running = true;

let arr = ["# #", "  #", " ##"];


let rows = len(arr);
let cols = len(arr[0]);
let i = 0;
let j = 0;
let out = ["   ", "   ", "   "];
let dead = true;
puts(arr[i][j] == "#");
puts("part");
if (arr[i][j] == "#") { dead = false; };

if (dead) {
    if (i == 1) { dead = false; };
}

while (i < rows) {
    while (j < cols) {
        let neighbours = 0;
        let up = i - 1;
        let down = i + 1;
        let left = j - 1;
        let right = j + 1;
        if (up < 0) { up = rows - 1 };
        if (down > rows - 1) { down = 0 };
        if (left < 0) {left = cols - 1};
        if (right > cols - 1) { right = 0};

        if (arr[up][left] == "#") { neighbours = neighbours + 1 };
        if (arr[up][j] == "#") { neighbours = neighbours + 1 };
        if (arr[up][right] == "#") { neighbours = neighbours + 1 };
        if (arr[i][left] == "#") { neighbours = neighbours + 1 };
        if (arr[i][right] == "#") { neighbours = neighbours + 1 };
        if (arr[down][left] == "#") { neighbours = neighbours + 1 };
        if (arr[down][j] == "#") { neighbours = neighbours + 1 };
        if (arr[down][right] == "#") { neighbours = neighbours + 1 };
        
        let dead = true;
        puts(arr[i][j] == "#");
        if (arr[i][j] == "#") { dead = false; };

        if (dead) {
            if (neighbours == 3) { dead = false; };
        } else {
            switch (neighbours) {
                case 2 { puts("reach"); dead = false; }
                case 3 { puts("reach1"); dead = false; }
                case 4 { puts("reach2"); dead = false; }
                case 5 { puts("reach3"); dead = false; }
                default { puts("reach4"); dead = true; }
            }
        }
        
        if (dead) {
            assign(out[i], j, " ");
        } else {
            assign(out[i], j, "#");
        }
        j = j + 1;
    }
    j = 0;
    i = i + 1;
}
