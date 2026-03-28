'use client'

import { NewAddress } from '@/app/restaurants/actions'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Address } from '@/lib/types'
import { Edit2, Loader2, Navigation, RotateCcw } from 'lucide-react'
import { useState } from 'react'

const EMPTY_ADDRESS = {
  addressName: '',
  subDistrict: '',
  district: '',
  province: '',
  postalCode: '',
}
const MAP_ZOOM = 18

const AddressPrompt = ({
  onConfirmAddress,
}: {
  onConfirmAddress: (addr: NewAddress) => Promise<Address>
}) => {
  const [loading, setLoading] = useState(false)
  const [hasDetected, setHasDetected] = useState(false)
  const [address, setAddress] = useState(EMPTY_ADDRESS)

  const handleGetLocation = () => {
    setLoading(true)

    setTimeout(() => {
      setAddress({
        addressName: '',
        subDistrict: 'Suthep',
        district: 'Mueang',
        province: 'Chiang Mai',
        postalCode: '50000',
      })
      setLoading(false)
      setHasDetected(true)
    }, 1200)
  }

  const handleReset = () => {
    setAddress(EMPTY_ADDRESS)
    setHasDetected(false)
  }

  const handleConfirmAddress = async () => {
    try {
      setLoading(true)
      await onConfirmAddress({
        addressName: address.addressName,
        district: address.district,
        postalCode: address.postalCode,
        province: address.province,
        subDistrict: address.subDistrict,
      })
    } catch (error) {
      console.error(error.message)
    } finally {
      setLoading(false)
    }
  }

  const { subDistrict, district, province, postalCode } = address
  const mapUrl = `https://maps.google.com/maps?q=${subDistrict},${district},${province},${postalCode}&z=${MAP_ZOOM}&output=embed`

  return (
    <div className="flex w-full items-center justify-center p-4">
      <Card
        className={`overflow-hidden transition-all duration-500 shadow-xl border-border ${hasDetected ? 'max-w-2xl' : 'w-fit border-dashed'
          }`}
      >
        <div
          className={`flex flex-col ${hasDetected ? 'md:flex-row max-h-full' : ''
            }`}
        >
          {hasDetected && <MapPreview mapUrl={mapUrl} />}

          <div
            className={`flex flex-col bg-card ${hasDetected ? 'md:w-[40%]' : 'w-full py-10'
              }`}
          >
            <CardContent className="flex-1 space-y-6">
              {!hasDetected ? (
                <DetectLocation
                  loading={loading}
                  onDetect={handleGetLocation}
                />
              ) : (
                <AddressForm
                  address={address}
                  setAddress={setAddress}
                  onReset={handleReset}
                />
              )}
            </CardContent>

            {hasDetected && (
              <CardFooter className="border-t pt-6 bg-muted/5">
                <Button
                  className="w-full font-bold shadow-md"
                  disabled={!address.addressName || loading}
                  onClick={handleConfirmAddress}
                >
                  {loading ? 'Saving...' : 'Confirm & Save'}
                </Button>
              </CardFooter>
            )}
          </div>
        </div>
      </Card>
    </div>
  )
}

const MapPreview = ({ mapUrl }: { mapUrl: string }) => {
  return (
    <div className="w-full md:w-[60%] bg-muted relative border-r">
      <iframe
        width="100%"
        height="100%"
        style={{ border: 0 }}
        src={mapUrl}
        title="Google Map"
        className="grayscale-20 contrast-[1.1]"
      />
    </div>
  )
}

const DetectLocation = ({ loading, onDetect, }: { loading: boolean, onDetect: () => void }) => {
  return (
    <div className="flex flex-col items-center justify-center">
      <p className="text-lg font-medium text-muted-foreground py-4">
        {loading
          ? 'Detecting...'
          : 'Please allow location access to see nearby restaurants.'}
      </p>

      <Button
        size="lg"
        className="h-16 w-16 rounded-full shadow-lg"
        onClick={onDetect}
        disabled={loading}
      >
        {loading ? (
          <Loader2 className="h-6 w-6 animate-spin" />
        ) : (
          <Navigation className="h-6 w-6" />
        )}
      </Button>
    </div>
  )
}

const AddressForm = ({
  address,
  setAddress,
  onReset,
}: {
  address: typeof EMPTY_ADDRESS
  setAddress: React.Dispatch<React.SetStateAction<typeof EMPTY_ADDRESS>>
  onReset: () => void
}) => {
  return (
    <div className="space-y-5 animate-in fade-in slide-in-from-right-4">
      <div className="space-y-2">
        <Label className="text-xs font-semibold text-primary inline-flex items-center gap-2">
          <Edit2 size={12} /> Place Label
        </Label>

        <Input
          placeholder="e.g. My Home"
          value={address.addressName}
          onChange={(e) =>
            setAddress((prev) => ({
              ...prev,
              addressName: e.target.value,
            }))
          }
          autoFocus
        />
      </div>

      <Separator />

      <div className="grid grid-cols-2 gap-4 opacity-70">
        <ReadOnlyField label="Sub-district" value={address.subDistrict} />
        <ReadOnlyField label="District" value={address.district} />
      </div>

      <div className="grid grid-cols-2 gap-4 opacity-70">
        <ReadOnlyField label="Province" value={address.province} />
        <ReadOnlyField label="Zip Code" value={address.postalCode} />
      </div>

      <Button
        variant="ghost"
        size="sm"
        className="w-full mb-4 text-blue-600"
        onClick={onReset}
      >
        <RotateCcw className="mr-2 h-3 w-3" /> Re-sync Location
      </Button>
    </div>
  )
}

const ReadOnlyField = ({ label, value }: { label: string; value: string }) => (
  <div className="space-y-1.5">
    <Label className="text-[11px] text-muted-foreground uppercase">
      {label}
    </Label>
    <Input
      value={value}
      readOnly
      className="h-9 bg-muted/50 border-none cursor-default"
    />
  </div>
)

export default AddressPrompt
