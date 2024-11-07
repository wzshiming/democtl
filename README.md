# democtl

democtl records terminal sessions and converts them into SVG animations and videos.
It creates visual demos of terminal interactions for documentation and presentations.

The tool is lightweight and standalone with no external dependencies.
Just install and start recording, then export to SVG animations.

## Key Features

- Record terminal sessions with .demo files that simulate typing
- No external dependencies required like browsers or runtimes
- Written in Go for cross-platform compatibility
- Compatible with asciinema .cast format 
- Export to multiple formats
  - SVG animations (no external dependencies)
  - MP4, GIF and WebM videos (with ffmpeg)

## Demo

[![color](https://github.com/wzshiming/democtl/raw/master/testdata/color.svg)](https://github.com/wzshiming/democtl/blob/master/testdata/color.demo)

[![base](https://github.com/wzshiming/democtl/raw/master/testdata/base.svg)](https://github.com/wzshiming/democtl/blob/master/testdata/base.demo)


## Installation

```bash
go install github.com/wzshiming/democtl/cmd/democtl@latest
```

## Usage

Record terminal commands to cast file.

```bash
democtl record --input ./testdata/base.demo --output ./testdata/base.cast
```

Convert cast file to svg file.

```bash
democtl svg --input ./testdata/base.cast --output ./testdata/base.svg
```

Convert cast file to video file.

```bash
democtl mp4 --input ./testdata/base.cast --output ./testdata/base.mp4
```

## Inspiration

[Originally written in shell script](https://github.com/wzshiming/democtl/blob/old/democtl.sh), democtl has been rewritten in Go for better maintainability and cross-platform support.

- [The old version](https://github.com/wzshiming/democtl/blob/old/democtl.sh)
- [asciinema](https://pypi.org/project/asciinema/)
- [playpty](https://pypi.org/project/playpty/)
- [svg-term-cli](https://www.npmjs.com/package/svg-term-cli)
- [@wzshiming/svg-to-video](https://www.npmjs.com/package/@wzshiming/svg-to-video)

## License

Licensed under the MIT License. See [LICENSE](https://github.com/wzshiming/democtl/blob/master/LICENSE) for the full license text.
