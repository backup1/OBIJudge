require('dotenv').load();
var gulp = require('gulp');
var csso = require('gulp-csso');
var concat = require('gulp-concat');
var minify = require('gulp-minify');
var sourcemaps = require("gulp-sourcemaps");
var htmlmin = require('gulp-htmlmin');
var googleWebFonts = require('gulp-google-webfonts')
var clean = require('gulp-clean');
var ifEnv = require('gulp-if-env');
var merge = require('merge-stream');
var sync = require('gulp-sync')(gulp).sync;
var reload = require('gulp-livereload');
var util = require('gulp-util');
var plumber = require('gulp-plumber');
var notifier = require('node-notifier');
var child = require('child_process');
var sudo = require('sudo');

var src = gulp.src;
gulp.src = function() {
  return src.apply(gulp, arguments)
    .pipe(plumber(function(error) {
      util.log(util.colors.red(
        'Error (' + error.plugin + '): ' + error.message
      ));
      notifier.notify({
        title: 'Error (' + error.plugin + ')',
        message: error.message.split('\n')[0]
      });
    }));
};

gulp.task('static:js', function() {
  return gulp.src([
      'node_modules/jquery/dist/jquery.js',
      'node_modules/katex/dist/katex.js',
      'node_modules/katex/dist/contrib/auto-render.js',
      'node_modules/clipboard/dist/clipboard.js',
      'node_modules/toastr/toastr.js',
      'node_modules/tippy.js/dist/tippy.js',
      'node_modules/moment/moment.js',

      'node_modules/codemirror/lib/codemirror.js',
      'node_modules/codemirror/addon/dialog/dialog.js',
      'node_modules/codemirror/addon/search/search.js',
      'node_modules/codemirror/addon/search/searchcursor.js',
      'node_modules/codemirror/addon/search/jump-to-line.js',
      'node_modules/codemirror/addon/search/match-highlighter.js',
      'node_modules/codemirror/addon/edit/matchbrackets.js',
      'node_modules/codemirror/addon/edit/closebrackets.js',
      'node_modules/codemirror/addon/display/placeholder.js',
      'node_modules/codemirror/addon/runmode/colorize.js',
      'node_modules/codemirror/addon/runmode/runmode.js',

      'node_modules/codemirror/mode/clike/clike.js',
      'node_modules/codemirror/mode/python/python.js',
      'node_modules/codemirror/mode/pascal/pascal.js',
      'node_modules/codemirror/mode/javascript/javascript.js',

      'static/src/*.js'
    ])
    .pipe(ifEnv.not('production', sourcemaps.init()))
    .pipe(concat('obijudge.js'))
    .pipe(ifEnv.not('production', sourcemaps.write()))
    .pipe(ifEnv('production', minify()))
    .pipe(gulp.dest('static/dist'));
})

gulp.task('static:css', function() {
  return gulp.src([
      'node_modules/normalize.css/normalize.css',
      'node_modules/skeleton-css/css/skeleton.css',
      'node_modules/katex/dist/katex.css',
      'node_modules/toastr/build/toastr.css',
      'node_modules/tippy.js/dist/tippy.css',
      'node_modules/tippy.js/dist/themes/light.css',
      'node_modules/codemirror/lib/codemirror.css',
      'node_modules/codemirror/addon/dialog/dialog.css',

      'static/src/*.css'
    ])
    .pipe(ifEnv.not('production', sourcemaps.init()))
    .pipe(ifEnv('production', csso({
      comments: false
    })))
    .pipe(concat('obijudge.css'))
    .pipe(ifEnv.not('production', sourcemaps.write()))
    .pipe(gulp.dest('static/dist'));
})

gulp.task('static:fonts', function() {
  var google = gulp.src('static/src/fonts.list')
    .pipe(googleWebFonts({
      fontsDir: 'fonts',
      cssDir: './',
      cssFilename: 'fonts.css',
      format: 'woff',
    }))
    .pipe(gulp.dest('static/dist'));

  var katex = gulp.src('node_modules/katex/dist/fonts/*.woff*')
    .pipe(gulp.dest('static/dist/fonts'));

  return merge(google, katex);
});

gulp.task('static:images', function() {
  return gulp.src(['static/src/obi.ico', 'static/src/obi.svg'])
    .pipe(gulp.dest('static/dist'))
});

gulp.task('static:templates', function() {
  return gulp.src([
      'templates/src/*.html'
    ])
    .pipe(ifEnv('production', htmlmin({
      caseSensitive: true,
      collapseWhitespace: true,
      ignoreCustomFragments: [
        /{{.*?}}/,
      ],
      minifyCSS: true,
      minifyJS: true,
      removeComments: true,
    })))
    .pipe(gulp.dest('templates/dist'));
})

gulp.task('static:build', ['static:js', 'static:css', 'static:fonts', 'static:images', 'static:templates']);

gulp.task('static:clean', function() {
  return gulp.src(['static/dist', 'templates/dist']).pipe(clean());
})

var server = null;
gulp.task('spawn', function() {
  if (server) {
    server.kill()
  }

  var cmd = './OBIJudge'
  var args = ['run']
  if (process.env.NODE_ENV !== 'production') {
    args.push('-testing')
  }
  server = child.spawn(cmd, args);

  server.stdout.once('data', function() {
    reload.reload('/');
  });

  server.stdout.on('data', function(data) {
    var lines = data.toString().split('\n')
    for (var l in lines)
      if (lines[l].length)
        util.log(lines[l]);
  });

  server.stderr.on('data', function(data) {
    process.stdout.write(data.toString());
  });
});

gulp.task('watch-spawn', function() {
  gulp.watch([
    'static/dist/*',
    'templates/dist/*',
    'locales/*',
    'OBIJudge',
    'contests.zip',
  ], ['spawn']);
})

gulp.task('watch-static', function() {
  gulp.watch([
    'static/src/*',
    'templates/src/*',
    'yarn.lock',
    'package.json',
  ], ['static:build']);
});
