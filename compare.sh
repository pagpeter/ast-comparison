rm -rf ./output && mkdir output
echo "[+] Getting test script"
curl -s https://cfschl.peet.ws/cdn-cgi/challenge-platform/h/g/orchestrate/chl_page/v1 > input.js 
echo "[+] Running js-babel"
node js-babel
echo "[+] Compiling & Running rust-swc"
cd rust-swc && cargo run --release -q && cd ..
echo "[+] Compiling & Running go-fAST"
cd go-fAST && go run . && cd ..