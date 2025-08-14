export type Download = {
    id: string;
    status: string;
    filename: string;
    url: string;
    source: string;
    progress: number;
    error: string | null;
    created_at: string;
    updated_at: string;
    speed: number;
    downloaded: number;
    start_time: string;
  };
