var fs = require('fs');
var rimraf = require('rimraf')
var archiver = require('archiver')

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

rimraf.sync('dist')

function build(buildCmd, cb) {
  result('git describe --tags --long', function(err, sha) {
    if (err) throw err;

    sha = sha.replace(/(.*)-(.*)/, '$1')
    sha = sha.replace(/^\s+|\s+$/g, '')

    const goos = process.env['GOOS']
    const goarch = process.env['GOARCH']

    console.log(`Building: ${sha} ${goos} ${goarch}`)

    process.env['VERSION'] = sha

    result(`yarn ${buildCmd}`, function(err, response) {
      if (err) throw err;

      console.log(response)

      if (goos == 'darwin') {
        var exeType = ''
      } else if (goos == 'windows') {
        var exeType = '.exe'
      }

      let exePath = 'gui-' + sha + exeType
      const destExePath = 'dist/' + exePath

      fs.mkdirSync('dist')

      fs.rename(`betaflight-pid-app${exeType}`, destExePath, function(err) {
        if (err) throw err;
        console.log(`Renamed app from betaflight-pid-app${exeType} to ` + destExePath)

        if (goos == 'darwin') {
          fs.mkdirSync('dist/Betaflight.app', 0o744)
          fs.mkdirSync('dist/Betaflight.app/Contents', 0o744)
          fs.mkdirSync('dist/Betaflight.app/Contents/MacOS', 0o744)
          fs.renameSync(destExePath, 'dist/Betaflight.app/Contents/MacOS/Betaflight')
          console.log('Packed Mac App ' + 'dist/Betaflight.app')
          var output = fs.createWriteStream('dist/Betaflight.zip')
          var archive = archiver('zip', {
            zlib: { level: 9 } // Sets the compression level.
          });
          archive.pipe(output)
          archive.directory('dist/Betaflight.app', 'Betaflight.app')
          archive.finalize()

          exePath = 'Betaflight.zip'
        }

        const versions = [
          {
            version: sha,
            file: exePath
          }
        ]

        fs.writeFile(`dist/${goos}-versions.json`, JSON.stringify(versions), function(err) {
          if (err) throw err;
          console.log(`Version: dist/${goos}-versions.json created`)
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