'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import NextImage from 'next/image';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'react-toastify';
import { z } from 'zod';

import { useIsMobileAndSubSection } from '@/app/setting/_hooks/isMobileAndSubSection';
import { AddressAvatar } from '@/components/address-avatar';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { InputAddon } from '@/components/ui/input-addon';
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import { useAccount } from '@/hooks/useAccount';

const formSchema = z.object({
  name: z.string().min(1, 'DAO name is required'),
  description: z.string().max(500, 'Description cannot exceed 500 characters'),
  daoUrl: z.string().min(1, 'DAO URL is required'),
  website: z.string().url('Please enter a valid URL').or(z.literal('')),
  twitter: z.string().optional(),
  discord: z.string().optional(),
  telegram: z.string().optional(),
  github: z.string().optional(),
  email: z.string().email('Invalid email').or(z.literal(''))
});

type FormValues = z.infer<typeof formSchema>;

const IMAGE_CONFIG = {
  maxWidth: 800,
  maxHeight: 800,
  quality: 0.7,
  acceptedFormats: ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml'],
  maxSizeMB: 10
};

export default function BasicSettingPage() {
  const { address } = useAccount();
  const { id } = useParams();
  const [isLoading, setIsLoading] = useState(false);
  const [avatar, setAvatar] = useState<string | null>(null);
  const [isAvatarUploading, setIsAvatarUploading] = useState(false);
  const isMobileAndSubSection = useIsMobileAndSubSection();

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      daoUrl: '',
      website: '',
      twitter: '',
      discord: '',
      telegram: '',
      github: '',
      email: ''
    }
  });

  async function onSubmit(values: FormValues) {
    setIsLoading(true);
    try {
      console.log('Form Values:', values);
      console.log('Avatar:', avatar);

      await new Promise((resolve) => setTimeout(resolve, 1000));

      toast.success('Settings saved successfully');
    } catch (error) {
      console.error('Error saving settings:', error);
      toast.error('Failed to save settings');
    } finally {
      setIsLoading(false);
    }
  }

  const compressImage = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();

      reader.onload = (event) => {
        const img = new Image();
        img.onload = () => {
          let width = img.width;
          let height = img.height;

          if (width > IMAGE_CONFIG.maxWidth) {
            height = (height * IMAGE_CONFIG.maxWidth) / width;
            width = IMAGE_CONFIG.maxWidth;
          }

          if (height > IMAGE_CONFIG.maxHeight) {
            width = (width * IMAGE_CONFIG.maxHeight) / height;
            height = IMAGE_CONFIG.maxHeight;
          }

          const canvas = document.createElement('canvas');
          canvas.width = width;
          canvas.height = height;

          const ctx = canvas.getContext('2d');
          if (!ctx) {
            reject(new Error('Failed to get canvas context'));
            return;
          }

          ctx.drawImage(img, 0, 0, width, height);

          const outputFormat = file.type === 'image/png' ? 'image/png' : 'image/jpeg';
          const base64 = canvas.toDataURL(outputFormat, IMAGE_CONFIG.quality);
          resolve(base64);
        };

        img.onerror = () => {
          reject(new Error('Failed to load image'));
        };

        img.src = event.target?.result as string;
      };

      reader.onerror = () => {
        reject(new Error('Failed to read file'));
      };

      reader.readAsDataURL(file);
    });
  };

  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!IMAGE_CONFIG.acceptedFormats.includes(file.type)) {
      toast.error('Please upload a JPG, PNG, GIF or WebP image.');
      return;
    }

    if (file.size > IMAGE_CONFIG.maxSizeMB * 1024 * 1024) {
      toast.error(`Maximum file size is ${IMAGE_CONFIG.maxSizeMB}MB.`);
      return;
    }

    try {
      setIsAvatarUploading(true);
      const base64 = await compressImage(file);
      setAvatar(base64);

      console.log('Avatar updated:', base64);
    } catch (error) {
      console.error('Error processing image:', error);
      toast.error('Image processing failed');
    } finally {
      setIsAvatarUploading(false);
    }
  };

  return (
    <div className="md:bg-card md:rounded-[14px] md:p-[20px]">
      {isMobileAndSubSection && (
        <Link href={`/setting/${id}`} className="flex items-center gap-[5px] md:gap-[10px]">
          <NextImage
            src="/back.svg"
            alt="back"
            width={32}
            height={32}
            className="size-[32px] flex-shrink-0"
          />
          <h1 className="text-[18px] font-semibold">Basic</h1>
        </Link>
      )}
      <div className="mt-[15px] flex items-start justify-center gap-[20px] md:mt-0">
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="space-y-[15px] md:space-y-[20px] md:p-[20px]"
          >
            <div className="space-y-[15px] md:space-y-[20px]">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-auto flex-shrink-0 text-[14px] md:w-[140px]">
                        DAO name
                      </FormLabel>
                      <FormControl>
                        <Input
                          placeholder="Enter your DAO name"
                          {...field}
                          className="w-full md:w-[410px]"
                        />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-auto flex-shrink-0 text-[14px] md:w-[140px]">
                        Description
                      </FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder="Write a description for your DAO. This will be displayed on the DAO dashboard"
                          className="min-h-[120px] w-full resize-none md:w-[410px]"
                          {...field}
                        />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="daoUrl"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-auto flex-shrink-0 text-[14px] md:w-[140px]">
                        DAO URL
                      </FormLabel>
                      <FormControl>
                        <InputAddon
                          suffix=".degov.ai"
                          placeholder="DAO-name"
                          suffixClassName="bg-muted text-muted-foreground"
                          containerClassName="w-full md:w-[410px]"
                          {...field}
                        />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="website"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-auto flex-shrink-0 text-[14px] md:w-[140px]">
                        Website
                      </FormLabel>
                      <FormControl>
                        <Input
                          placeholder="The DAO site"
                          {...field}
                          className="w-full md:w-[410px]"
                        />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="twitter"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-auto flex-shrink-0 text-[14px] md:w-[140px]">
                        Twitter
                      </FormLabel>
                      <FormControl>
                        <Input placeholder="@Username" {...field} className="w-full md:w-[410px]" />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="discord"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-auto flex-shrink-0 text-[14px] md:w-[140px]">
                        Discord
                      </FormLabel>
                      <FormControl>
                        <Input placeholder="Username" {...field} className="w-full md:w-[410px]" />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="telegram"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-auto flex-shrink-0 text-[14px] md:w-[140px]">
                        Telegram
                      </FormLabel>
                      <FormControl>
                        <Input placeholder="@Username" {...field} className="w-full md:w-[410px]" />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="github"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-auto flex-shrink-0 text-[14px] md:w-[140px]">
                        Github
                      </FormLabel>
                      <FormControl>
                        <Input placeholder="Username" {...field} className="w-full md:w-[410px]" />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex flex-col gap-[10px] md:flex-row md:items-center">
                      <FormLabel className="w-[140px] flex-shrink-0 text-[14px]">Email</FormLabel>
                      <FormControl>
                        <Input
                          type="email"
                          placeholder="Email@example.com"
                          {...field}
                          className="w-full md:w-[410px]"
                        />
                      </FormControl>
                    </div>
                    <div className="pl-0 md:pl-[160px]">
                      <FormMessage />
                    </div>
                  </FormItem>
                )}
              />
            </div>
            <Separator className="hidden md:my-[20px] md:block" />
            <div className="grid grid-cols-[1fr_1fr] gap-[20px] md:flex md:justify-center">
              <Button
                type="button"
                variant="outline"
                className="w-auto rounded-full p-[10px] md:w-[155px]"
                onClick={() => form.reset()}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                className="w-auto rounded-full p-[10px] md:w-[155px]"
                disabled={isLoading}
              >
                {isLoading ? 'Saving...' : 'Save'}
              </Button>
            </div>
          </form>
        </Form>

        <div className="hidden flex-col items-center gap-[20px] p-[20px] md:flex">
          <div className="relative h-[110px] w-[110px] overflow-hidden rounded-full">
            {avatar ? (
              <NextImage
                src={avatar}
                alt="DAO avatar"
                width={110}
                height={110}
                className="h-full w-full object-cover"
              />
            ) : (
              address && (
                <AddressAvatar
                  address={address}
                  className="h-[110px] w-[110px] rounded-full"
                  size={110}
                />
              )
            )}

            {isAvatarUploading && (
              <div className="absolute inset-0 flex items-center justify-center bg-black/50">
                <div className="h-8 w-8 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
              </div>
            )}
          </div>

          <input
            type="file"
            id="avatar-upload"
            onChange={handleAvatarChange}
            accept={IMAGE_CONFIG.acceptedFormats.join(',')}
            className="hidden"
          />

          <Button
            type="button"
            variant="outline"
            className="w-full rounded-[14px] border"
            onClick={() => document.getElementById('avatar-upload')?.click()}
            disabled={isAvatarUploading}
          >
            {isAvatarUploading ? 'Uploading...' : 'Edit'}
          </Button>
        </div>
      </div>
    </div>
  );
}
