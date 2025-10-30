export default function (archives: any) {
  for (const arch of archives) {
    if (arch.cover_path == null || arch.cover_path == "") {
      $fetch(`/api/a/${arch.id}/cover`, {
        method: "POST",
      });
    }
  }
};