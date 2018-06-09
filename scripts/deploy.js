// set SHA=$(git rev-parse --short HEAD); yarn build-386 && mv betaflight-pid-app.exe dist/gui2-$SHA.exe

var fs = require('fs');

var exec = require('child_process').exec;

var result = function(command, cb){
  var child = exec(command, function(err, stdout, stderr){
    if(err != null){
        return cb(new Error(err), null);
    }else if(typeof(stderr) != "string"){
        return cb(new Error(stderr), null);
    }else{
        return cb(null, stdout);
    }
  });
}

const build = require('./build').build

const buildCmd = process.argv[2]

if (!buildCmd) {
  console.error("Missing build command")
  process.exit(-1)
}

build(buildCmd, function(sha) {
  result(`scp -r dist/. foo.us.to:www/foo.us.to/html/gui2/`, function(err) {
    if (err) {
      console.error(err)
      return
    }

    console.log(`Deployed ${sha} to foo.us.to`)
  })
})