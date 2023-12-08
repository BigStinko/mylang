let populate = func(rows, cols) {
    let arr = [];
    let i = 0;
    let j = 0;
    while (i < rows) {
        let str = "";
        while (j < cols) {
            let box = " ";
            if (rand() > 0.5) {
                box = "#";
            }
            str = str + box;
            j = j + 1;
        }
        j = 0;
        arr = push(arr, str);
        i = i + 1;
    }
    return arr;
}

let solve = func(arr) {
    let rows = len(arr);
    let cols = len(arr[0]);
    let i = 0;
    let j = 0;
    let out = populate(rows, cols);

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
            if (arr[i][j] == "#") { dead = false; };

            if (dead) {
                if (neighbours == 3) { dead = false; };
            } else {
                switch (neighbours) {
                    case 2 { dead = false; }
                    case 3 { dead = false; }
                    default { dead = true; }
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
    return out;
}

let printArr = func(arr) {
    let i = 0;
    while (i < len(arr)) {
        puts(arr[i]);
        i = i + 1;
    }
}

let running = true;
let input = open("/dev/stdin");
let clear = command("clear")["stdout"];

while (running) {
    puts("enter x dimension");
    let str = read(input);
    pop(str);
    let cols = int(str);
    puts("enter y dimension");
    str = read(input);
    pop(str);
    let rows = int(str);
    
    puts(clear);
    let arr = populate(rows, cols);
    printArr(arr);
    
    let solving = true;
    while (solving) {
        puts("press enter to iterate and anything else to stop");
        let stop = read(input);
        puts(clear);
        if (stop != "\n") { 
            solving = false;
        } else {
            arr = solve(arr);
            printArr(arr);
        }
    }

    puts("quit program, y or n?");
    let quit = read(input);
    pop(quit);
    if (quit == "y") {
        running = false;
    }
}
