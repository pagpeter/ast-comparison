echo "[+] Running js-babel"
time node js-babel
echo "[+] Running rust-swc"
cd rust-swc && time cargo run && cd ..
echo "[+] Running go-fAST"
cd go-fAST && time go run . && cd ..