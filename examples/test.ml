let global = 1;

let fn = func() {
    let a = 2;
    a = 10;

    func() {
        let b = 3;

        func() {
            let c = 4;

            global + a + b + c;
        }
    }
}

puts(fn()()());
puts(global);
