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

function build(buildCmd, cb) {
  result('git describe --tags --long', function(err, sha) {
    sha = sha.replace(/(.*)-(.*)/, '$1')
    sha = sha.replace(/^\s+|\s+$/g, '')

    const goos = process.env['GOOS']
    const goarch = process.env['GOARCH']

    console.log(`Building: ${sha} ${goos} ${goarch}`)

    process.env['VERSION'] = sha

    result(`yarn ${buildCmd}`, function(err, response) {
      if (err) throw err;

      console.log(response)

      const exePath = 'gui-' + sha + '.exe'
      const destExePath = 'dist/' + exePath

      fs.rename('betaflight-pid-app.exe', destExePath, function(err) {
        if (err) throw err;
        console.log('Renamed app from betaflight-pid-app.exe to ' + destExePath)

        const versions = [
          {
            version: sha,
            file: exePath
          }
        ]

        fs.writeFile('dist/versions.json', JSON.stringify(versions), function(err) {
          if (err) throw err;
          console.log('Version: dist/versions.json created')
        })

        if (typeof(cb) === 'function') {
          cb(sha)
        }
      })
    })
  })
}

if (require.main === module) {
  const buildCmd = process.argv[2]

  if (!buildCmd) {
    console.error("Missing build command")
    process.exit(-1)
  }

  build(buildCmd)
}

exports.build = build