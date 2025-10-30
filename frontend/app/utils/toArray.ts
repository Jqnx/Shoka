export default function (input: any[]) {
  const output: string[] = [];
  if (!input) {
    return undefined;
  } else {
    for (const i of input) {
      if (i.name) {
        output.push(i.name);
      } else if (i.url) {
        output.push(i.url);
      } else if (i.artist) {
        output.push(i.artist);
      } else if (i.tag) {
        output.push(i.tag);
      } else if (i.parody) {
        output.push(i.parody);
      } else if (i.character) {
        output.push(i.character);
      }
    }
    return output;
  }
}
